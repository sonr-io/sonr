package integration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/bridge/handlers"
	"github.com/sonr-io/sonr/crypto/ucan"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

// perfTestDIDKeeper implements DIDKeeperInterface for performance tests
type perfTestDIDKeeper struct {
	didDocuments map[string]*didtypes.DIDDocument
}

func (m *perfTestDIDKeeper) GetDIDDocument(ctx context.Context, did string) (*didtypes.DIDDocument, error) {
	if doc, ok := m.didDocuments[did]; ok {
		return doc, nil
	}
	return nil, fmt.Errorf("DID document not found: %s", did)
}

func (m *perfTestDIDKeeper) GetVerificationMethod(ctx context.Context, did string, methodID string) (*didtypes.VerificationMethod, error) {
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

// createTestDIDKeeper creates a mock DID keeper with test DIDs
func createTestDIDKeeper() *perfTestDIDKeeper {
	pubKey, _, _ := ed25519.GenerateKey(rand.Reader)

	return &perfTestDIDKeeper{
		didDocuments: map[string]*didtypes.DIDDocument{
			"did:sonr:test-user": {
				Id: "did:sonr:test-user",
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						Id:                     "did:sonr:test-user#keys-1",
						Controller:             "did:sonr:test-user",
						VerificationMethodKind: "Ed25519VerificationKey2020",
						PublicKeyMultibase:     base64.StdEncoding.EncodeToString(pubKey),
					},
				},
			},
			"did:sonr:test-client": {
				Id: "did:sonr:test-client",
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						Id:                     "did:sonr:test-client#keys-1",
						Controller:             "did:sonr:test-client",
						VerificationMethodKind: "Ed25519VerificationKey2020",
						PublicKeyMultibase:     base64.StdEncoding.EncodeToString(pubKey),
					},
				},
			},
		},
	}
}

// BenchmarkUCANCreation benchmarks UCAN token creation performance
func BenchmarkUCANCreation(b *testing.B) {
	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")
	delegator := handlers.NewUCANDelegator(signer)

	userDID := "did:sonr:bench-user"
	clientDID := "did:sonr:bench-client"
	scopes := []string{"vault:read", "dwn:write", "service:read"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := delegator.CreateDelegation(
			userDID,
			clientDID,
			scopes,
			time.Now().Add(time.Hour),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUCANSigning benchmarks UCAN token signing performance
func BenchmarkUCANSigning(b *testing.B) {
	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")

	// Create template token
	token := &ucan.Token{
		Issuer:    "did:sonr:bench-user",
		Audience:  "did:sonr:bench-client",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		NotBefore: time.Now().Unix(),
		Attenuations: []ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := signer.Sign(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUCANVerification benchmarks UCAN token verification performance
func BenchmarkUCANVerification(b *testing.B) {
	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")

	// Create test token
	token := &ucan.Token{
		Issuer:    "did:sonr:bench-user",
		Audience:  "did:sonr:bench-client",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		NotBefore: time.Now().Unix(),
		Attenuations: []ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
	}

	signedToken, err := signer.Sign(token)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := signer.VerifySignature(signedToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOAuth2ScopeMapping benchmarks scope to UCAN mapping performance
func BenchmarkOAuth2ScopeMapping(b *testing.B) {
	scopeMapper := handlers.NewScopeMapper()

	scopes := []string{"openid", "profile", "vault:read", "vault:write", "dwn:read", "dwn:write", "service:manage"}
	userDID := "did:sonr:bench-user"
	clientID := "bench-client"
	resourceContext := map[string]string{
		"vault_address": "vault:test",
		"dwn_id":        "dwn:test",
		"service_id":    "service:test",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		attenuations := scopeMapper.MapToUCAN(scopes, userDID, clientID, resourceContext)
		if len(attenuations) == 0 {
			b.Fatal("No attenuations generated")
		}
	}
}

// BenchmarkDelegationChainValidation benchmarks delegation chain validation
func BenchmarkDelegationChainValidation(b *testing.B) {
	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")

	// Create delegation chain: User → Client A → Client B → Client C
	userDID := "did:sonr:bench-user"
	clientA := "did:sonr:bench-client-a"
	clientB := "did:sonr:bench-client-b"
	clientC := "did:sonr:bench-client-c"

	token1, _ := signer.CreateDelegationToken(
		userDID,
		clientA,
		[]ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{Actions: []string{"read", "write"}},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: userDID},
			},
		},
		nil,
		time.Hour,
	)

	token2, _ := signer.CreateDelegationToken(
		clientA,
		clientB,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: userDID},
			},
		},
		[]ucan.Proof{ucan.Proof(token1)},
		time.Hour,
	)

	token3, _ := signer.CreateDelegationToken(
		clientB,
		clientC,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: userDID},
			},
		},
		[]ucan.Proof{ucan.Proof(token2)},
		time.Hour,
	)

	tokenChain := []string{token1, token2, token3}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := signer.ValidateDelegationChain(tokenChain)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTokenExchange benchmarks OAuth2 token exchange performance
func BenchmarkTokenExchange(b *testing.B) {
	clientStore := NewMockClientStore()
	tokenStore := NewMockTokenStore()

	// Setup test client
	clientStore.clients["bench-client"] = &handlers.OAuth2Client{
		ClientID:      "bench-client",
		ClientSecret:  "bench-secret",
		AllowedScopes: []string{"vault:read", "dwn:write"},
		Metadata: map[string]string{
			"client_did": "did:sonr:bench-client",
		},
	}

	// Setup existing access token
	tokenStore.StoreToken(context.Background(), &handlers.StoredToken{
		TokenID:     "bench-access-token",
		TokenType:   "access_token",
		AccessToken: "bench-access-token",
		ClientID:    "bench-client",
		UserDID:     "did:sonr:bench-user",
		Scopes:      []string{"vault:read", "dwn:write"},
		ExpiresAt:   time.Now().Add(time.Hour),
		UCANToken:   "dummy.ucan.token",
	})

	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")
	delegator := handlers.NewUCANDelegator(signer)
	exchanger := handlers.NewTokenExchangeHandler(delegator, signer, tokenStore, clientStore)

	req := &handlers.TokenExchangeRequest{
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		SubjectToken:       "bench-access-token",
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:access_token",
		RequestedTokenType: "urn:x-oath:params:oauth:token-type:ucan",
		Scope:              "vault:read",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create HTTP request/response for benchmarking
		reqBody := fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:token-exchange&subject_token=%s&subject_token_type=urn:ietf:params:oauth:token-type:access_token&client_id=test-client", req.SubjectToken)
		httpReq := httptest.NewRequest("POST", "/oauth2/token-exchange", strings.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		recorder := httptest.NewRecorder()
		exchanger.HandleTokenExchange(recorder, httpReq)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Token exchange failed: %d", recorder.Code)
		}
	}
}

// BenchmarkRefreshToken benchmarks refresh token performance
func BenchmarkRefreshToken(b *testing.B) {
	clientStore := NewMockClientStore()
	tokenStore := NewMockTokenStore()

	// Setup test client
	clientStore.clients["bench-client"] = &handlers.OAuth2Client{
		ClientID:      "bench-client",
		ClientSecret:  "bench-secret",
		AllowedScopes: []string{"vault:read", "dwn:write"},
		Metadata: map[string]string{
			"client_did": "did:sonr:bench-client",
		},
	}

	// Setup refresh token
	tokenStore.StoreToken(context.Background(), &handlers.StoredToken{
		TokenID:      "bench-refresh-token",
		TokenType:    "refresh_token",
		RefreshToken: "bench-refresh-token",
		ClientID:     "bench-client",
		UserDID:      "did:sonr:bench-user",
		Scopes:       []string{"vault:read", "dwn:write"},
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	})

	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")
	delegator := handlers.NewUCANDelegator(signer)
	refresher := handlers.NewRefreshTokenHandler(delegator, signer, tokenStore, clientStore)

	req := &handlers.RefreshTokenRequest{
		GrantType:    "refresh_token",
		RefreshToken: "bench-refresh-token",
		Scope:        "vault:read",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create HTTP request for benchmarking
		reqBody := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s&client_id=bench-client", req.RefreshToken)
		httpReq := httptest.NewRequest("POST", "/oauth2/refresh", strings.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		recorder := httptest.NewRecorder()
		refresher.HandleRefreshToken(recorder, httpReq)

		if recorder.Code != http.StatusOK {
			b.Fatalf("Refresh token failed: %d", recorder.Code)
		}
	}
}

// BenchmarkConcurrentUCANOperations benchmarks concurrent UCAN operations
func BenchmarkConcurrentUCANOperations(b *testing.B) {
	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")
	delegator := handlers.NewUCANDelegator(signer)

	userDID := "did:sonr:concurrent-user"
	clientDID := "did:sonr:concurrent-client"
	scopes := []string{"vault:read", "dwn:write"}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			token, err := delegator.CreateDelegation(
				userDID,
				clientDID,
				scopes,
				time.Now().Add(time.Hour),
			)
			if err != nil {
				b.Fatal(err)
			}

			// Verify the token
			_, err = signer.VerifySignature(token.Raw)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkLargeAttenuationList benchmarks performance with many attenuations
func BenchmarkLargeAttenuationList(b *testing.B) {
	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")

	// Create token with many attenuations
	var attenuations []ucan.Attenuation
	modules := []string{"vault", "dwn", "service", "did", "dex"}
	actions := []string{"read", "write", "delete", "admin"}

	for _, module := range modules {
		for _, action := range actions {
			attenuations = append(attenuations, ucan.Attenuation{
				Capability: &ucan.SimpleCapability{Action: action},
				Resource:   &handlers.SimpleResource{Scheme: module, Value: "test"},
			})
		}
	}

	token := &ucan.Token{
		Issuer:       "did:sonr:large-user",
		Audience:     "did:sonr:large-client",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		signedToken, err := signer.Sign(token)
		if err != nil {
			b.Fatal(err)
		}

		_, err = signer.VerifySignature(signedToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkModuleSpecificValidation benchmarks module-specific validation
func BenchmarkModuleSpecificValidation(b *testing.B) {
	didValidator := NewMockDIDValidator()
	dwnValidator := NewMockDWNValidator()
	svcValidator := NewMockServiceValidator()

	// Setup test data
	didValidator.didDocuments["did:sonr:bench-user"] = &didtypes.DIDDocument{
		Id: "did:sonr:bench-user",
	}

	svcValidator.services["bench-service"] = &svctypes.Service{
		Id:    "bench-service",
		Owner: "did:sonr:bench-user",
	}

	signer, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")
	delegator := handlers.NewUCANDelegator(signer)

	token, _ := delegator.CreateDelegation(
		"did:sonr:bench-user",
		"did:sonr:bench-client",
		[]string{"did:read", "dwn:write", "service:read"},
		time.Now().Add(time.Hour),
	)

	b.ResetTimer()
	b.ReportAllocs()

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		// Test all three modules
		_ = didValidator.ValidateUCANPermission(ctx, token, "read", "did:sonr:bench-user")
		_ = dwnValidator.ValidateUCANPermission(ctx, token, "write", "dwn:bench-user")
		_ = svcValidator.ValidateUCANPermission(ctx, token, "read", "bench-service")
	}
}

// Performance test with target metrics
func TestPerformanceTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	// Define performance targets
	targets := map[string]time.Duration{
		"UCAN Creation":     10 * time.Millisecond,
		"UCAN Signing":      5 * time.Millisecond,
		"UCAN Verification": 5 * time.Millisecond,
		"Scope Mapping":     1 * time.Millisecond,
		"Chain Validation":  15 * time.Millisecond,
	}

	results := make(map[string]time.Duration)

	// Measure UCAN Creation
	start := time.Now()
	iterations := 1000

	mockKeeper := createTestDIDKeeper()
	signer, _ := handlers.NewBlockchainUCANSigner(mockKeeper, "did:sonr:oauth-provider")
	delegator := handlers.NewUCANDelegator(signer)

	for i := 0; i < iterations; i++ {
		_, err := delegator.CreateDelegation(
			"did:sonr:test-user",
			"did:sonr:test-client",
			[]string{"vault:read", "dwn:write"},
			time.Now().Add(time.Hour),
		)
		require.NoError(t, err)
	}
	results["UCAN Creation"] = time.Since(start) / time.Duration(iterations)

	// Measure UCAN Signing
	token := &ucan.Token{
		Issuer:    "did:sonr:test-user",
		Audience:  "did:sonr:test-client",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Attenuations: []ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
	}

	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := signer.Sign(token)
		require.NoError(t, err)
	}
	results["UCAN Signing"] = time.Since(start) / time.Duration(iterations)

	// Measure UCAN Verification
	signedToken, err := signer.Sign(token)
	require.NoError(t, err)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := signer.VerifySignature(signedToken)
		require.NoError(t, err)
	}
	results["UCAN Verification"] = time.Since(start) / time.Duration(iterations)

	// Measure Scope Mapping
	scopeMapper := handlers.NewScopeMapper()
	scopes := []string{"vault:read", "dwn:write", "service:read"}
	resourceContext := map[string]string{"vault_address": "test"}

	start = time.Now()
	for i := 0; i < iterations; i++ {
		attenuations := scopeMapper.MapToUCAN(scopes, "user", "client", resourceContext)
		require.NotEmpty(t, attenuations)
	}
	results["Scope Mapping"] = time.Since(start) / time.Duration(iterations)

	// Measure Chain Validation
	token1, _ := signer.CreateDelegationToken(
		"did:sonr:user",
		"did:sonr:client-a",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)

	token2, _ := signer.CreateDelegationToken(
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

	chain := []string{token1, token2}

	start = time.Now()
	for i := 0; i < iterations; i++ {
		err := signer.ValidateDelegationChain(chain)
		require.NoError(t, err)
	}
	results["Chain Validation"] = time.Since(start) / time.Duration(iterations)

	// Check against targets
	t.Logf("Performance Results:")
	for operation, target := range targets {
		actual := results[operation]
		t.Logf("  %s: %v (target: %v)", operation, actual, target)

		if actual > target {
			t.Errorf("%s performance (%v) exceeds target (%v)", operation, actual, target)
		}
	}

	// Memory usage test
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	before := memStats.Alloc

	// Perform operations
	for i := 0; i < 100; i++ {
		token, _ := delegator.CreateDelegation(
			"did:sonr:mem-user",
			"did:sonr:mem-client",
			[]string{"vault:read"},
			time.Now().Add(time.Hour),
		)
		_, _ = signer.Sign(token)
	}

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	after := memStats.Alloc

	memUsed := after - before
	t.Logf("Memory used for 100 operations: %d bytes (%d KB)", memUsed, memUsed/1024)

	// Target: < 1MB for 100 operations
	if memUsed > 1024*1024 {
		t.Errorf("Memory usage (%d bytes) exceeds 1MB target", memUsed)
	}
}
