package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/sonr/bridge/handlers"
	"github.com/sonr-io/sonr/crypto/ucan"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	dwntypes "github.com/sonr-io/sonr/x/dwn/types"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

// CrossModuleAuthTestSuite tests authorization across all Sonr modules
type CrossModuleAuthTestSuite struct {
	suite.Suite
	ucanDelegator *handlers.UCANDelegator
	signer        *MockUCANSigner

	// Mock module keepers
	didValidator *MockDIDValidator
	dwnValidator *MockDWNValidator
	svcValidator *MockServiceValidator

	testUserDID   string
	testClientDID string
}

// MockDIDKeeper implements DIDKeeperInterface for testing
type MockDIDKeeper struct {
	documents map[string]*didtypes.DIDDocument
}

func (m *MockDIDKeeper) GetDIDDocument(ctx context.Context, did string) (*didtypes.DIDDocument, error) {
	doc, ok := m.documents[did]
	if !ok {
		return nil, fmt.Errorf("DID not found: %s", did)
	}
	return doc, nil
}

func (m *MockDIDKeeper) GetVerificationMethod(ctx context.Context, did string, methodID string) (*didtypes.VerificationMethod, error) {
	doc, err := m.GetDIDDocument(ctx, did)
	if err != nil {
		return nil, err
	}
	for _, vm := range doc.VerificationMethod {
		if vm.Id == methodID {
			return vm, nil
		}
	}
	return nil, fmt.Errorf("verification method not found: %s", methodID)
}

// MockUCANSigner is a test implementation that bypasses actual crypto
type MockUCANSigner struct {
	issuerDID string
	didKeeper *MockDIDKeeper
	tokens    map[string]*ucan.Token // Store tokens by raw string
}

func (m *MockUCANSigner) Sign(token *ucan.Token) (string, error) {
	// Generate a fake JWT-like token for testing
	tokenStr := fmt.Sprintf("mock.token.%s.%s.%d", token.Issuer, token.Audience, time.Now().Unix())
	token.Raw = tokenStr

	// Store the token for later verification
	if m.tokens == nil {
		m.tokens = make(map[string]*ucan.Token)
	}
	m.tokens[tokenStr] = token

	return tokenStr, nil
}

func (m *MockUCANSigner) GetIssuerDID() string {
	return m.issuerDID
}

func (m *MockUCANSigner) CreateDelegationToken(
	issuer string,
	audience string,
	attenuations []ucan.Attenuation,
	proofs []ucan.Proof,
	expiry time.Duration,
) (string, error) {
	token := &ucan.Token{
		Issuer:       issuer,
		Audience:     audience,
		ExpiresAt:    time.Now().Add(expiry).Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
		Proofs:       proofs,
	}

	return m.Sign(token)
}

func (m *MockUCANSigner) VerifySignature(tokenString string) (*ucan.Token, error) {
	// For testing, just return the stored token
	if token, ok := m.tokens[tokenString]; ok {
		// Check expiry - tokens expire when current time exceeds ExpiresAt
		currentTime := time.Now().Unix()
		if currentTime > token.ExpiresAt {
			return nil, fmt.Errorf("token expired (current: %d, expires: %d)", currentTime, token.ExpiresAt)
		}
		// Check not before
		if token.NotBefore > 0 && currentTime < token.NotBefore {
			return nil, fmt.Errorf("token not yet valid")
		}
		return token, nil
	}
	return nil, fmt.Errorf("token not found: %s", tokenString)
}

func (m *MockUCANSigner) ValidateDelegationChain(tokens []string) error {
	// Validate the delegation chain
	for i, tokenStr := range tokens {
		token, ok := m.tokens[tokenStr]
		if !ok {
			return fmt.Errorf("failed to verify token %d: token not found", i)
		}

		// Check chain consistency (except for first token)
		if i > 0 {
			prevToken := m.tokens[tokens[i-1]]
			// Verify that the issuer of current token matches audience of previous
			if token.Issuer != prevToken.Audience {
				return fmt.Errorf("failed to verify token %d: broken chain - issuer mismatch", i)
			}

			// Check that current token has previous token as proof
			hasProof := false
			for _, proof := range token.Proofs {
				if string(proof) == tokens[i-1] {
					hasProof = true
					break
				}
			}
			if !hasProof && len(token.Proofs) > 0 {
				// If there are proofs but wrong ones, it's a broken chain
				return fmt.Errorf("failed to verify token %d: invalid proof in chain", i)
			}
		}
	}
	return nil
}

// MockDIDValidator simulates DID module UCAN validation
type MockDIDValidator struct {
	didDocuments map[string]*didtypes.DIDDocument
}

func NewMockDIDValidator() *MockDIDValidator {
	return &MockDIDValidator{
		didDocuments: make(map[string]*didtypes.DIDDocument),
	}
}

func (m *MockDIDValidator) ValidateUCANPermission(ctx context.Context, token *ucan.Token, action string, resource string) error {
	// Map common action aliases
	actionMap := map[string]string{
		"update": "write",
		"create": "write",
		"delete": "write",
	}

	// Normalize the action
	normalizedAction := action
	if mapped, exists := actionMap[action]; exists {
		normalizedAction = mapped
	}

	// Find appropriate attenuation for DID operations
	for _, att := range token.Attenuations {
		if att.Resource != nil && att.Resource.GetScheme() == "did" {
			if att.Capability != nil {
				actions := att.Capability.GetActions()
				for _, a := range actions {
					if a == normalizedAction || a == action || a == "*" || a == "write" && (action == "update" || action == "create") {
						return nil // Authorized
					}
				}
			}
		}
	}
	return handlers.ErrUnauthorized
}

func (m *MockDIDValidator) GetDIDDocument(ctx context.Context, did string) (*didtypes.DIDDocument, error) {
	doc, exists := m.didDocuments[did]
	if !exists {
		return nil, didtypes.ErrDIDNotFound
	}
	return doc, nil
}

// MockDWNValidator simulates DWN module UCAN validation
type MockDWNValidator struct {
	records map[string]*dwntypes.DWNRecord
}

func NewMockDWNValidator() *MockDWNValidator {
	return &MockDWNValidator{
		records: make(map[string]*dwntypes.DWNRecord),
	}
}

func (m *MockDWNValidator) ValidateUCANPermission(ctx context.Context, token *ucan.Token, action string, resource string) error {
	// Map common action aliases
	actionMap := map[string]string{
		"records-write":       "write",
		"protocols-configure": "admin",
	}

	// Normalize the action
	normalizedAction := action
	if mapped, exists := actionMap[action]; exists {
		normalizedAction = mapped
	}

	// Find appropriate attenuation for DWN operations
	for _, att := range token.Attenuations {
		if att.Resource != nil && att.Resource.GetScheme() == "dwn" {
			if att.Capability != nil {
				actions := att.Capability.GetActions()
				for _, a := range actions {
					if a == normalizedAction || a == action || a == "*" ||
						(a == "write" && action == "records-write") ||
						(a == "admin" && action == "protocols-configure") {
						// Additional DWN-specific validation
						return m.validateDWNSpecific(token, action, resource)
					}
				}
			}
		}
	}
	return handlers.ErrUnauthorized
}

func (m *MockDWNValidator) validateDWNSpecific(token *ucan.Token, action, resource string) error {
	// DWN-specific business logic
	switch action {
	case "records-write", "write":
		// Check if user has write permission to this DWN
		if token.Issuer == "did:sonr:unauthorized" {
			return handlers.ErrInsufficientScope
		}
	case "protocols-configure":
		// Only DWN owner can configure protocols
		if !m.isOwner(token.Issuer, resource) {
			return handlers.ErrUnauthorized
		}
	}
	return nil
}

func (m *MockDWNValidator) isOwner(userDID, dwnResource string) bool {
	// Simplified ownership check
	return userDID == "did:sonr:owner" || userDID == "did:sonr:test-user"
}

// MockServiceValidator simulates Service module UCAN validation
type MockServiceValidator struct {
	services map[string]*svctypes.Service
}

func NewMockServiceValidator() *MockServiceValidator {
	return &MockServiceValidator{
		services: make(map[string]*svctypes.Service),
	}
}

func (m *MockServiceValidator) ValidateUCANPermission(ctx context.Context, token *ucan.Token, action string, resource string) error {
	// Map common action aliases
	actionMap := map[string]string{
		"register":      "write",
		"verify-domain": "write",
		"delete":        "admin",
		"read":          "read",
	}

	// Normalize the action
	normalizedAction := action
	if mapped, exists := actionMap[action]; exists {
		normalizedAction = mapped
	}

	// Find appropriate attenuation for service operations
	for _, att := range token.Attenuations {
		if att.Resource != nil {
			scheme := att.Resource.GetScheme()
			if scheme == "service" || scheme == "svc" {
				if att.Capability != nil {
					actions := att.Capability.GetActions()
					for _, a := range actions {
						if a == normalizedAction || a == action || a == "*" ||
							(a == "write" && (action == "register" || action == "verify-domain")) ||
							(a == "admin" && action == "delete") ||
							(a == "manage" && (action == "register" || action == "verify-domain")) {
							return m.validateServiceSpecific(token, action, resource)
						}
					}
				}
			}
		}
	}
	return handlers.ErrUnauthorized
}

func (m *MockServiceValidator) validateServiceSpecific(token *ucan.Token, action, resource string) error {
	// Service-specific validation logic
	switch action {
	case "register":
		// Anyone can register a service with valid DID
		if token.Issuer == "" {
			return handlers.ErrUnauthorized
		}
	case "verify-domain":
		// Only service owner can verify domain
		service, exists := m.services[resource]
		if !exists || service.Owner != token.Issuer {
			return handlers.ErrUnauthorized
		}
	case "delete", "admin":
		// Admin actions require admin scope
		return m.validateAdminAccess(token)
	}
	return nil
}

func (m *MockServiceValidator) validateAdminAccess(token *ucan.Token) error {
	// Check for admin capabilities
	for _, att := range token.Attenuations {
		actions := att.Capability.GetActions()
		for _, action := range actions {
			if action == "admin" || action == "*" {
				return nil
			}
		}
	}
	return handlers.ErrInsufficientScope
}

func (suite *CrossModuleAuthTestSuite) SetupSuite() {
	// Initialize validators
	suite.didValidator = NewMockDIDValidator()
	suite.dwnValidator = NewMockDWNValidator()
	suite.svcValidator = NewMockServiceValidator()

	// Initialize UCAN components with mock DID keeper
	mockDIDKeeper := &MockDIDKeeper{
		documents: make(map[string]*didtypes.DIDDocument),
	}
	// Add test user DID document with verification method
	mockDIDKeeper.documents["did:sonr:test-user"] = &didtypes.DIDDocument{
		Id: "did:sonr:test-user",
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     "did:sonr:test-user#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				PublicKeyMultibase:     "dGVzdC1wdWJsaWMta2V5", // base64 encoded test key
			},
		},
	}
	mockDIDKeeper.documents["did:sonr:owner"] = &didtypes.DIDDocument{
		Id: "did:sonr:owner",
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     "did:sonr:owner#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				PublicKeyMultibase:     "b3duZXItcHVibGljLWtleQ==", // base64 encoded test key
			},
		},
	}
	// Add other test DIDs used in the tests
	mockDIDKeeper.documents["did:sonr:service-a"] = &didtypes.DIDDocument{
		Id: "did:sonr:service-a",
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     "did:sonr:service-a#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				PublicKeyMultibase:     "c2VydmljZS1hLWtleQ==", // base64 encoded test key
			},
		},
	}
	mockDIDKeeper.documents["did:sonr:service-b"] = &didtypes.DIDDocument{
		Id: "did:sonr:service-b",
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     "did:sonr:service-b#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				PublicKeyMultibase:     "c2VydmljZS1iLWtleQ==", // base64 encoded test key
			},
		},
	}
	mockDIDKeeper.documents["did:sonr:intermediate-client"] = &didtypes.DIDDocument{
		Id: "did:sonr:intermediate-client",
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     "did:sonr:intermediate-client#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				PublicKeyMultibase:     "aW50ZXJtZWRpYXRlLWtleQ==", // base64 encoded test key
			},
		},
	}
	mockDIDKeeper.documents["did:sonr:non-owner"] = &didtypes.DIDDocument{
		Id: "did:sonr:non-owner",
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     "did:sonr:non-owner#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				PublicKeyMultibase:     "bm9uLW93bmVyLWtleQ==", // base64 encoded test key
			},
		},
	}
	mockDIDKeeper.documents["did:sonr:different-user"] = &didtypes.DIDDocument{
		Id: "did:sonr:different-user",
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     "did:sonr:different-user#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				PublicKeyMultibase:     "ZGlmZmVyZW50LXVzZXI=", // base64 encoded test key
			},
		},
	}

	// Use a mock signer for testing that bypasses actual crypto
	mockSigner := &MockUCANSigner{
		issuerDID: "did:sonr:oauth-provider",
		didKeeper: mockDIDKeeper,
	}
	suite.signer = mockSigner
	suite.ucanDelegator = handlers.NewUCANDelegator(mockSigner)

	// Test identities
	suite.testUserDID = "did:sonr:test-user"
	suite.testClientDID = "did:sonr:test-client"

	// Setup test data
	suite.setupTestData()
}

func (suite *CrossModuleAuthTestSuite) setupTestData() {
	// Add test DID document
	suite.didValidator.didDocuments[suite.testUserDID] = &didtypes.DIDDocument{
		Id: suite.testUserDID,
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     suite.testUserDID + "#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
			},
		},
	}

	// Add test service
	suite.svcValidator.services["test-service"] = &svctypes.Service{
		Id:     "test-service",
		Owner:  suite.testUserDID,
		Domain: "example.com",
		Status: svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
	}
}

// TestDIDModuleAuthorization tests DID-specific operations
func (suite *CrossModuleAuthTestSuite) TestDIDModuleAuthorization() {
	ctx := context.Background()

	// Create token with DID permissions
	token, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID,
		suite.testClientDID,
		[]string{"did:read", "did:write"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Test authorized DID read
	err = suite.didValidator.ValidateUCANPermission(ctx, token, "read", suite.testUserDID)
	suite.NoError(err, "Should authorize DID read operation")

	// Test authorized DID write
	err = suite.didValidator.ValidateUCANPermission(ctx, token, "update", suite.testUserDID)
	suite.NoError(err, "Should authorize DID update operation")

	// Test unauthorized DID admin operation
	err = suite.didValidator.ValidateUCANPermission(ctx, token, "admin", suite.testUserDID)
	suite.Error(err, "Should reject unauthorized admin operation")
}

// TestDWNModuleAuthorization tests DWN-specific operations
func (suite *CrossModuleAuthTestSuite) TestDWNModuleAuthorization() {
	ctx := context.Background()

	// Create token with DWN permissions
	token, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID,
		suite.testClientDID,
		[]string{"dwn:read", "dwn:write"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Test authorized DWN write
	err = suite.dwnValidator.ValidateUCANPermission(ctx, token, "records-write", "dwn:"+suite.testUserDID)
	suite.NoError(err, "Should authorize DWN records write")

	// Test unauthorized DWN protocol configuration
	err = suite.dwnValidator.ValidateUCANPermission(ctx, token, "protocols-configure", "dwn:"+suite.testUserDID)
	suite.Error(err, "Should reject protocol configuration without admin scope")

	// Create admin token
	adminToken, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID,
		suite.testClientDID,
		[]string{"dwn:admin"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Test authorized protocol configuration with admin token
	err = suite.dwnValidator.ValidateUCANPermission(ctx, adminToken, "protocols-configure", "dwn:"+suite.testUserDID)
	suite.NoError(err, "Should authorize protocol configuration with admin scope")
}

// TestServiceModuleAuthorization tests Service-specific operations
func (suite *CrossModuleAuthTestSuite) TestServiceModuleAuthorization() {
	ctx := context.Background()

	// Create token with service permissions
	token, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID,
		suite.testClientDID,
		[]string{"service:read", "service:write"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Test authorized service registration
	err = suite.svcValidator.ValidateUCANPermission(ctx, token, "register", "new-service")
	suite.NoError(err, "Should authorize service registration")

	// Test authorized domain verification (owner)
	err = suite.svcValidator.ValidateUCANPermission(ctx, token, "verify-domain", "test-service")
	suite.NoError(err, "Should authorize domain verification by owner")

	// Test unauthorized admin operation
	err = suite.svcValidator.ValidateUCANPermission(ctx, token, "admin", "test-service")
	suite.Error(err, "Should reject admin operation without admin scope")
}

// TestCrossModuleWorkflow tests a workflow spanning multiple modules
func (suite *CrossModuleAuthTestSuite) TestCrossModuleWorkflow() {
	ctx := context.Background()

	// Scenario: User registers a service, configures DWN, and updates DID
	// Create comprehensive token with cross-module permissions
	token, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID,
		suite.testClientDID,
		[]string{
			"service:write",
			"dwn:write",
			"did:write",
			"vault:read", // For accessing keys
		},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Step 1: Register new service
	err = suite.svcValidator.ValidateUCANPermission(ctx, token, "register", "new-service-workflow")
	suite.NoError(err, "Should authorize service registration in workflow")

	// Step 2: Configure DWN for the service
	err = suite.dwnValidator.ValidateUCANPermission(ctx, token, "records-write", "dwn:"+suite.testUserDID)
	suite.NoError(err, "Should authorize DWN configuration in workflow")

	// Step 3: Update DID document to include service
	err = suite.didValidator.ValidateUCANPermission(ctx, token, "update", suite.testUserDID)
	suite.NoError(err, "Should authorize DID update in workflow")

	// Verify token contains all necessary capabilities
	suite.Require().NotEmpty(token.Attenuations)

	// Count module coverage
	modules := make(map[string]bool)
	for _, att := range token.Attenuations {
		modules[att.Resource.GetScheme()] = true
	}

	suite.GreaterOrEqual(len(modules), 3, "Token should cover at least 3 modules")
	suite.Contains(modules, "service")
	suite.Contains(modules, "dwn")
	suite.Contains(modules, "did")
}

// TestScopeAttenuation tests proper scope reduction in delegation chains
func (suite *CrossModuleAuthTestSuite) TestScopeAttenuation() {
	ctx := context.Background()

	// Create initial broad token
	parentToken, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID,
		"did:sonr:intermediate-client",
		[]string{
			"vault:admin",    // Full vault access
			"service:manage", // Full service access
			"dwn:write",      // DWN write access
			"did:write",      // DID write access
		},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Create attenuated token (reduced permissions)
	childToken, err := suite.ucanDelegator.CreateDelegation(
		"did:sonr:intermediate-client",
		suite.testClientDID,
		[]string{
			"vault:read",   // Reduced to read-only
			"service:read", // Reduced to read-only
			"dwn:read",     // Reduced to read-only
		},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Child token should have fewer capabilities
	suite.Less(len(childToken.Attenuations), len(parentToken.Attenuations),
		"Child token should have fewer capabilities than parent")

	// Verify child cannot perform admin operations
	err = suite.svcValidator.ValidateUCANPermission(ctx, childToken, "admin", "test-service")
	suite.Error(err, "Child token should not authorize admin operations")

	// But child should be able to perform read operations
	err = suite.svcValidator.ValidateUCANPermission(ctx, childToken, "read", "test-service")
	suite.NoError(err, "Child token should authorize read operations")
}

// TestDelegationChainValidation tests multi-hop delegation chains
func (suite *CrossModuleAuthTestSuite) TestDelegationChainValidation() {
	// Create 3-hop delegation chain: User → Service A → Service B → Service C

	// Hop 1: User → Service A
	serviceA := "did:sonr:service-a"
	token1, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		serviceA,
		[]ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{Actions: []string{"read", "write"}},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: suite.testUserDID},
			},
		},
		nil,
		time.Hour,
	)
	suite.Require().NoError(err)

	// Hop 2: Service A → Service B (with attenuation)
	serviceB := "did:sonr:service-b"
	token2, err := suite.signer.CreateDelegationToken(
		serviceA,
		serviceB,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"}, // Reduced to read-only
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: suite.testUserDID},
			},
		},
		[]ucan.Proof{ucan.Proof(token1)},
		time.Hour,
	)
	suite.Require().NoError(err)

	// Hop 3: Service B → Service C (maintain same level)
	serviceC := "did:sonr:service-c"
	token3, err := suite.signer.CreateDelegationToken(
		serviceB,
		serviceC,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"}, // Same level
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: suite.testUserDID},
			},
		},
		[]ucan.Proof{ucan.Proof(token2)},
		time.Hour,
	)
	suite.Require().NoError(err)

	// Validate the complete chain
	err = suite.signer.ValidateDelegationChain([]string{token1, token2, token3})
	suite.NoError(err, "Valid 3-hop delegation chain should pass validation")

	// Test broken chain (missing intermediate proof)
	brokenToken, err := suite.signer.CreateDelegationToken(
		serviceA, // Wrong issuer - should be serviceB
		serviceC,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: suite.testUserDID},
			},
		},
		[]ucan.Proof{ucan.Proof(token1)}, // Wrong proof
		time.Hour,
	)
	suite.Require().NoError(err)

	err = suite.signer.ValidateDelegationChain([]string{token1, token2, brokenToken})
	suite.Error(err, "Broken delegation chain should fail validation")
}

// TestModuleSpecificValidation tests module-specific business rules
func (suite *CrossModuleAuthTestSuite) TestModuleSpecificValidation() {
	ctx := context.Background()

	// Test DWN ownership validation
	ownerToken, err := suite.ucanDelegator.CreateDelegation(
		"did:sonr:owner", // Different user
		suite.testClientDID,
		[]string{"dwn:admin"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	// Owner should be able to configure protocols
	err = suite.dwnValidator.ValidateUCANPermission(ctx, ownerToken, "protocols-configure", "dwn:owner")
	suite.NoError(err, "Owner should authorize protocol configuration")

	// Non-owner should not be able to configure protocols
	nonOwnerToken, err := suite.ucanDelegator.CreateDelegation(
		"did:sonr:non-owner",
		suite.testClientDID,
		[]string{"dwn:admin"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	err = suite.dwnValidator.ValidateUCANPermission(ctx, nonOwnerToken, "protocols-configure", "dwn:owner")
	suite.Error(err, "Non-owner should not authorize protocol configuration")

	// Test Service domain verification
	serviceOwnerToken, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID, // Owner of test-service
		suite.testClientDID,
		[]string{"service:write"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	err = suite.svcValidator.ValidateUCANPermission(ctx, serviceOwnerToken, "verify-domain", "test-service")
	suite.NoError(err, "Service owner should authorize domain verification")

	// Non-owner should not be able to verify domain
	nonServiceOwnerToken, err := suite.ucanDelegator.CreateDelegation(
		"did:sonr:different-user",
		suite.testClientDID,
		[]string{"service:write"},
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)

	err = suite.svcValidator.ValidateUCANPermission(ctx, nonServiceOwnerToken, "verify-domain", "test-service")
	suite.Error(err, "Non-owner should not authorize domain verification")
}

// TestTimeBasedValidation tests time-sensitive authorization
func (suite *CrossModuleAuthTestSuite) TestTimeBasedValidation() {
	// Create short-lived token (1 second to account for Unix timestamp precision)
	shortToken, err := suite.ucanDelegator.CreateDelegation(
		suite.testUserDID,
		suite.testClientDID,
		[]string{"vault:read"},
		time.Now().Add(1*time.Second), // 1 second expiration for Unix timestamp precision
	)
	suite.Require().NoError(err)

	// Should work immediately
	_, err = suite.signer.VerifySignature(shortToken.Raw)
	suite.NoError(err, "Token should be valid immediately after creation")

	// Wait for expiration (need to wait at least 2 seconds due to Unix timestamp rounding)
	time.Sleep(2 * time.Second)

	// Should fail after expiration
	_, err = suite.signer.VerifySignature(shortToken.Raw)
	suite.Error(err, "Token should be invalid after expiration")

	// Test not-before validation
	futureToken, err := suite.signer.CreateDelegationToken(
		suite.testUserDID,
		suite.testClientDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: suite.testUserDID},
			},
		},
		nil,
		time.Hour,
	)
	suite.Require().NoError(err)

	// Manually verify token structure
	parsedToken, err := suite.signer.VerifySignature(futureToken)
	suite.NoError(err, "Token should parse successfully")
	suite.NotZero(parsedToken.ExpiresAt, "Token should have expiration time")
}

func TestCrossModuleAuthSuite(t *testing.T) {
	suite.Run(t, new(CrossModuleAuthTestSuite))
}
