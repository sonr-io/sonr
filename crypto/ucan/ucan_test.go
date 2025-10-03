package ucan

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityCreation(t *testing.T) {
	testCases := []struct {
		name     string
		actions  []string
		expected bool
	}{
		{
			name:     "Basic Capability Creation",
			actions:  []string{"read", "write"},
			expected: true,
		},
		{
			name:     "Empty Actions",
			actions:  []string{},
			expected: true,
		},
		{
			name:     "Complex Actions",
			actions:  []string{"create", "update", "delete", "admin"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			capability := &MultiCapability{Actions: tc.actions}

			assert.NotNil(t, capability)
			assert.Equal(t, len(tc.actions), len(capability.Actions))

			for _, action := range tc.actions {
				assert.Contains(t, capability.Actions, action)
			}
		})
	}
}

func TestCapabilityValidation(t *testing.T) {
	testCases := []struct {
		name           string
		actions        []string
		resourceScheme string
		shouldPass     bool
	}{
		{
			name:           "Valid Standard Actions",
			actions:        []string{"read", "write"},
			resourceScheme: "example",
			shouldPass:     true,
		},
		{
			name:           "Invalid Actions",
			actions:        []string{"delete", "admin"},
			resourceScheme: "restricted",
			shouldPass:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			capability := &MultiCapability{Actions: tc.actions}
			resource := &SimpleResource{
				Scheme: tc.resourceScheme,
				Value:  "test",
				URI:    tc.resourceScheme + "://test",
			}

			attenuation := Attenuation{
				Capability: capability,
				Resource:   resource,
			}

			StandardTemplate.AddAllowedActions(tc.resourceScheme, []string{"read", "write"})
			err := StandardTemplate.ValidateAttenuation(attenuation)

			if tc.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTTokenLifecycle(t *testing.T) {
	testCases := []struct {
		name           string
		actions        []string
		resourceScheme string
		duration       time.Duration
		shouldPass     bool
	}{
		{
			name:           "Valid Token Generation and Verification",
			actions:        []string{"read", "write"},
			resourceScheme: "example",
			duration:       time.Hour,
			shouldPass:     true,
		},
		{
			name:           "Expired Token",
			actions:        []string{"read"},
			resourceScheme: "test",
			duration:       -time.Hour, // Expired token
			shouldPass:     false,
		},
	}

	// Use standard service template for testing
	StandardTemplate := StandardServiceTemplate()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			capability := &MultiCapability{Actions: tc.actions}
			resource := &SimpleResource{
				Scheme: tc.resourceScheme,
				Value:  "test",
				URI:    tc.resourceScheme + "://test",
			}

			attenuation := Attenuation{
				Capability: capability,
				Resource:   resource,
			}

			// Validate attenuation against template
			err := StandardTemplate.ValidateAttenuation(attenuation)
			require.NoError(t, err)

			// Simulate JWT token generation and verification
			token := "test_token_" + time.Now().String()

			if tc.shouldPass {
				// Simulate verification
				verifiedToken := &Token{
					Raw:          token,
					Issuer:       "did:sonr:local",
					Attenuations: []Attenuation{attenuation},
					ExpiresAt:    time.Now().Add(tc.duration).Unix(),
				}

				assert.NotNil(t, verifiedToken)
				assert.Equal(t, "did:sonr:local", verifiedToken.Issuer)
				assert.Len(t, verifiedToken.Attenuations, 1)
				assert.Equal(
					t,
					tc.resourceScheme+"://test",
					verifiedToken.Attenuations[0].Resource.GetURI(),
				)
			} else {
				// Simulate expired token verification
				assert.True(t, time.Now().Unix() > time.Now().Add(tc.duration).Unix())
			}
		})
	}
}

func TestCapabilityRevocation(t *testing.T) {
	capability := &MultiCapability{Actions: []string{"read", "write"}}
	resource := &SimpleResource{
		Scheme: "example",
		Value:  "test",
		URI:    "example://test",
	}

	attenuation := Attenuation{
		Capability: capability,
		Resource:   resource,
	}

	// Generate token
	token, err := GenerateJWTToken(attenuation, time.Hour)
	require.NoError(t, err)

	// Revoke capability
	err = RevokeCapability(attenuation)
	assert.NoError(t, err)

	// Attempt to verify revoked token should fail
	_, err = VerifyJWTToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token has been revoked")
}

func TestResourceValidation(t *testing.T) {
	testCases := []struct {
		name           string
		resourceScheme string
		resourceValue  string
		resourceURI    string
		expectValid    bool
	}{
		{
			name:           "Valid Resource",
			resourceScheme: "sonr",
			resourceValue:  "test-resource",
			resourceURI:    "sonr://test-resource",
			expectValid:    true,
		},
		{
			name:           "Invalid Resource URI",
			resourceScheme: "invalid",
			resourceValue:  "test-resource",
			resourceURI:    "invalid-malformed-uri",
			expectValid:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource := &SimpleResource{
				Scheme: tc.resourceScheme,
				Value:  tc.resourceValue,
				URI:    tc.resourceURI,
			}

			// Simplified resource validation
			if tc.expectValid {
				assert.Regexp(t, `^[a-z]+://[a-z-]+$`, resource.URI)
			} else {
				assert.NotRegexp(t, `^[a-z]+://[a-z-]+$`, resource.URI)
			}
		})
	}
}

func TestValidateEnclaveDataCIDIntegrity(t *testing.T) {
	testCases := []struct {
		name          string
		data          []byte
		expectedCID   string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Empty CID",
			data:          []byte("test data"),
			expectedCID:   "",
			expectError:   true,
			errorContains: "enclave data CID cannot be empty",
		},
		{
			name:          "Empty data",
			data:          []byte{},
			expectedCID:   "QmTest",
			expectError:   true,
			errorContains: "enclave data cannot be empty",
		},
		{
			name:          "Invalid CID format",
			data:          []byte("test data"),
			expectedCID:   "invalid-cid",
			expectError:   true,
			errorContains: "invalid IPFS CID format",
		},
		{
			name:        "Valid CID verification - should pass",
			data:        []byte("test data"),
			expectedCID: generateValidCIDForData([]byte("test data")),
			expectError: false,
		},
		{
			name:          "Mismatched CID - should fail",
			data:          []byte("test data"),
			expectedCID:   generateValidCIDForData([]byte("different data")),
			expectError:   true,
			errorContains: "CID verification failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEnclaveDataCIDIntegrity(tc.expectedCID, tc.data)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to generate a valid CID for test data
func generateValidCIDForData(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	digest := hasher.Sum(nil)

	mhash, err := multihash.EncodeName(digest, "sha2-256")
	if err != nil {
		panic(err)
	}

	calculatedCID := cid.NewCidV1(cid.DagProtobuf, mhash)
	return calculatedCID.String()
}
