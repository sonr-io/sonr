package keeper

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/sonr-io/sonr/api/svc/v1"
)

// Domain verification constants
const (
	// VerificationPrefix is the prefix for DNS TXT records
	VerificationPrefix = "sonr-verification="

	// TokenLength is the length of the verification token in bytes
	TokenLength = 32

	// VerificationExpiryHours is how long a verification token is valid
	VerificationExpiryHours = 24
)

// InitiateDomainVerification creates a new domain verification request
func (k Keeper) InitiateDomainVerification(
	ctx context.Context,
	domain, owner string,
) (*v1.DomainVerification, error) {
	// Validate domain format
	if err := k.validateDomainFormat(domain); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid domain format: %v", err)
	}

	// Check if domain verification already exists and is not expired
	existing, err := k.OrmDB.DomainVerificationTable().Get(ctx, domain)
	if err == nil {
		// Domain verification exists, check if it's still valid
		if k.isDomainVerificationValid(existing) {
			return existing, status.Errorf(
				codes.AlreadyExists,
				"domain verification already exists and is valid",
			)
		}

		// Expired verification exists, we'll update it
	}

	// Generate a new verification token
	token, err := k.generateVerificationToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate verification token: %v", err)
	}

	// Create new domain verification record
	now := time.Now().Unix()
	verification := &v1.DomainVerification{
		Domain:            domain,
		Owner:             owner,
		VerificationToken: token,
		Status:            v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_PENDING,
		ExpiresAt:         now + (VerificationExpiryHours * 3600), // 24 hours from now
		VerifiedAt:        0,
	}

	// Save or update the verification record
	if existing != nil {
		// Update existing record
		verification.Domain = existing.Domain // Ensure primary key consistency
		err = k.OrmDB.DomainVerificationTable().Update(ctx, verification)
	} else {
		// Insert new record
		err = k.OrmDB.DomainVerificationTable().Insert(ctx, verification)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save domain verification: %v", err)
	}

	return verification, nil
}

// VerifyDomainOwnership validates domain ownership by checking DNS TXT records
func (k Keeper) VerifyDomainOwnership(
	ctx context.Context,
	domain string,
) (*v1.DomainVerification, error) {
	// Get the domain verification record
	verification, err := k.OrmDB.DomainVerificationTable().Get(ctx, domain)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "domain verification not found: %v", err)
	}

	// Check if verification has expired
	if k.isDomainVerificationExpired(verification) {
		verification.Status = v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_EXPIRED
		k.OrmDB.DomainVerificationTable().Update(ctx, verification)
		return verification, status.Errorf(
			codes.DeadlineExceeded,
			"domain verification has expired",
		)
	}

	// Check if already verified
	if verification.Status == v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED {
		return verification, nil
	}

	// Perform DNS TXT record lookup
	verified, err := k.checkDNSTXTRecord(domain, verification.VerificationToken)
	if err != nil {
		verification.Status = v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_FAILED
		k.OrmDB.DomainVerificationTable().Update(ctx, verification)
		return verification, status.Errorf(
			codes.FailedPrecondition,
			"DNS verification failed: %v",
			err,
		)
	}

	if verified {
		// Mark as verified
		verification.Status = v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED
		verification.VerifiedAt = time.Now().Unix()
	} else {
		// Verification record not found
		verification.Status = v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_FAILED
	}

	// Update the verification record
	err = k.OrmDB.DomainVerificationTable().Update(ctx, verification)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update domain verification: %v", err)
	}

	if !verified {
		return verification, status.Errorf(
			codes.FailedPrecondition,
			"verification record not found in DNS",
		)
	}

	return verification, nil
}

// GetDomainVerification retrieves a domain verification record
func (k Keeper) GetDomainVerification(
	ctx context.Context,
	domain string,
) (*v1.DomainVerification, error) {
	verification, err := k.OrmDB.DomainVerificationTable().Get(ctx, domain)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "domain verification not found: %v", err)
	}
	return verification, nil
}

// ListDomainVerificationsByOwner returns all domain verifications for a given owner
func (k Keeper) ListDomainVerificationsByOwner(
	ctx context.Context,
	owner string,
) ([]*v1.DomainVerification, error) {
	ownerKey := v1.DomainVerificationOwnerIndexKey{}.WithOwner(owner)
	iter, err := k.OrmDB.DomainVerificationTable().List(ctx, ownerKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list domain verifications: %v", err)
	}
	defer iter.Close()

	var verifications []*v1.DomainVerification
	for iter.Next() {
		verification, err := iter.Value()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to read domain verification: %v", err)
		}
		verifications = append(verifications, verification)
	}

	return verifications, nil
}

// IsVerifiedDomain checks if a domain is verified and not expired
func (k Keeper) IsVerifiedDomain(ctx context.Context, domain string) bool {
	verification, err := k.OrmDB.DomainVerificationTable().Get(ctx, domain)
	if err != nil {
		return false
	}

	return verification.Status == v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED &&
		!k.isDomainVerificationExpired(verification)
}

// generateVerificationToken creates a cryptographically secure random token
func (k Keeper) generateVerificationToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// checkDNSTXTRecord performs DNS TXT record lookup and validation
func (k Keeper) checkDNSTXTRecord(domain, expectedToken string) (bool, error) {
	// Expected TXT record format: "sonr-verification=<token>"
	expectedRecord := VerificationPrefix + expectedToken

	// Perform DNS TXT lookup
	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		return false, fmt.Errorf("DNS lookup failed: %w", err)
	}

	// Check if any TXT record matches our expected verification record
	for _, record := range txtRecords {
		if strings.TrimSpace(record) == expectedRecord {
			return true, nil
		}
	}

	return false, nil
}

// validateDomainFormat validates that a domain name is properly formatted
func (k Keeper) validateDomainFormat(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Basic domain validation - check for valid characters and format
	if len(domain) > 253 {
		return fmt.Errorf("domain name too long")
	}

	// Check for valid domain format (basic validation)
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("domain must contain at least one dot")
	}

	// Check for invalid characters
	for _, char := range domain {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '-') {
			return fmt.Errorf("domain contains invalid character: %c", char)
		}
	}

	// Domain cannot start or end with a hyphen
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return fmt.Errorf("domain cannot start or end with hyphen")
	}

	return nil
}

// isDomainVerificationValid checks if a domain verification is still valid (not expired)
func (k Keeper) isDomainVerificationValid(verification *v1.DomainVerification) bool {
	if verification.Status == v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED {
		return true // Verified domains don't expire
	}

	return !k.isDomainVerificationExpired(verification)
}

// isDomainVerificationExpired checks if a domain verification has expired
func (k Keeper) isDomainVerificationExpired(verification *v1.DomainVerification) bool {
	now := time.Now().Unix()
	return now > verification.ExpiresAt
}

// GetDNSInstructions returns human-readable instructions for setting up DNS verification
func (k Keeper) GetDNSInstructions(domain, token string) string {
	return fmt.Sprintf(
		"Add the following TXT record to your DNS configuration for domain '%s':\n\n"+
			"Name: %s\n"+
			"Type: TXT\n"+
			"Value: %s%s\n\n"+
			"Note: DNS propagation may take up to 48 hours. You can verify the record using:\n"+
			"dig TXT %s",
		domain, domain, VerificationPrefix, token, domain,
	)
}

// SetDomainVerified is a helper method for testing to mark a domain as verified
func (k Keeper) SetDomainVerified(ctx context.Context, domain string) error {
	verification, err := k.OrmDB.DomainVerificationTable().Get(ctx, domain)
	if err != nil {
		return fmt.Errorf("domain verification not found: %w", err)
	}

	verification.Status = v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED
	return k.OrmDB.DomainVerificationTable().Update(ctx, verification)
}
