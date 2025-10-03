package test

import (
	"testing"
	"time"

	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleSpecificUCANCapabilities(t *testing.T) {
	t.Run("TestDIDCapabilityCreation", func(t *testing.T) {
		// Create DID-specific attenuation
		didAttenuation := ucan.CreateDIDAttenuation(
			[]string{"create", "update"},
			"*",
			[]string{"owner"},
		)

		assert.Equal(t, "did:*", didAttenuation.Resource.GetURI())

		didCap, ok := didAttenuation.Capability.(*ucan.DIDCapability)
		require.True(t, ok, "Expected DIDCapability")
		assert.Contains(t, didCap.Actions, "create")
		assert.Contains(t, didCap.Actions, "update")
		assert.Contains(t, didCap.Caveats, "owner")
	})

	t.Run("TestDWNCapabilityCreation", func(t *testing.T) {
		// Create DWN-specific attenuation
		dwnAttenuation := ucan.CreateDWNAttenuation(
			[]string{"create", "read", "update", "delete"},
			"personal/*",
			[]string{"owner"},
		)

		assert.Equal(t, "dwn:records/personal/*", dwnAttenuation.Resource.GetURI())

		dwnCap, ok := dwnAttenuation.Capability.(*ucan.DWNCapability)
		require.True(t, ok, "Expected DWNCapability")
		assert.Contains(t, dwnCap.Actions, "create")
		assert.Contains(t, dwnCap.Actions, "read")
		assert.Contains(t, dwnCap.Caveats, "owner")
	})

	t.Run("TestDEXCapabilityCreation", func(t *testing.T) {
		// Create DEX-specific attenuation
		dexAttenuation := ucan.CreateDEXAttenuation(
			[]string{"swap", "provide-liquidity"},
			"snr/usd",
			[]string{"max-amount"},
			"1000snr",
		)

		assert.Equal(t, "dex:pool/snr/usd", dexAttenuation.Resource.GetURI())

		dexCap, ok := dexAttenuation.Capability.(*ucan.DEXCapability)
		require.True(t, ok, "Expected DEXCapability")
		assert.Contains(t, dexCap.Actions, "swap")
		assert.Contains(t, dexCap.Actions, "provide-liquidity")
		assert.Contains(t, dexCap.Caveats, "max-amount")
		assert.Equal(t, "1000snr", dexCap.MaxAmount)
	})

	t.Run("TestCrossModuleCapability", func(t *testing.T) {
		// Create cross-module capability
		crossCap := &ucan.CrossModuleCapability{
			Modules: map[string]ucan.Capability{
				"did": &ucan.DIDCapability{
					Actions: []string{"create", "update"},
				},
				"dwn": &ucan.DWNCapability{
					Actions: []string{"create", "read"},
				},
			},
		}

		actions := crossCap.GetActions()
		assert.Contains(t, actions, "create")
		assert.Contains(t, actions, "update")
		assert.Contains(t, actions, "read")

		assert.True(t, crossCap.Grants([]string{"create"}))
		assert.True(t, crossCap.Grants([]string{"update", "read"}))
		assert.False(t, crossCap.Grants([]string{"delete"}))
	})

	t.Run("TestGaslessCapability", func(t *testing.T) {
		// Create gasless capability wrapper
		baseCap := &ucan.DIDCapability{
			Actions: []string{"create"},
		}

		gaslessCap := &ucan.GaslessCapability{
			Capability:   baseCap,
			AllowGasless: true,
			GasLimit:     100000,
		}

		assert.True(t, gaslessCap.SupportsGasless())
		assert.Equal(t, uint64(100000), gaslessCap.GetGasLimit())
		assert.True(t, gaslessCap.Grants([]string{"create"}))
	})
}

func TestModuleJWTTokenGeneration(t *testing.T) {
	t.Run("TestGenerateAndVerifyModuleToken", func(t *testing.T) {
		// Create attenuations for different modules
		attenuations := []ucan.Attenuation{
			ucan.CreateDIDAttenuation([]string{"create", "update"}, "*", []string{"owner"}),
			ucan.CreateDWNAttenuation([]string{"create", "read"}, "personal/*", []string{"owner"}),
		}

		issuer := "did:key:alice"
		audience := "did:key:bob"

		// Generate token
		tokenString, err := ucan.GenerateModuleJWTToken(
			attenuations,
			issuer,
			audience,
			time.Hour,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		// Verify token
		token, err := ucan.VerifyModuleJWTToken(tokenString, issuer, audience)
		require.NoError(t, err)
		assert.Equal(t, issuer, token.Issuer)
		assert.Equal(t, audience, token.Audience)
		assert.Len(t, token.Attenuations, 2)

		// Check that attenuations are properly parsed
		var didAtt, dwnAtt *ucan.Attenuation
		for _, att := range token.Attenuations {
			switch att.Resource.GetScheme() {
			case "did":
				didAtt = &att
			case "dwn":
				dwnAtt = &att
			}
		}

		require.NotNil(t, didAtt, "DID attenuation not found")
		require.NotNil(t, dwnAtt, "DWN attenuation not found")

		didCap, ok := didAtt.Capability.(*ucan.DIDCapability)
		assert.True(t, ok)
		assert.Contains(t, didCap.Actions, "create")
		assert.Contains(t, didCap.Actions, "update")

		dwnCap, ok := dwnAtt.Capability.(*ucan.DWNCapability)
		assert.True(t, ok)
		assert.Contains(t, dwnCap.Actions, "create")
		assert.Contains(t, dwnCap.Actions, "read")
	})
}

func TestModuleCapabilityTemplates(t *testing.T) {
	t.Run("TestDIDTemplate", func(t *testing.T) {
		template := ucan.StandardDIDTemplate()

		// Test valid DID actions
		didAtt := ucan.CreateDIDAttenuation([]string{"create", "update"}, "*", nil)
		err := template.ValidateAttenuation(didAtt)
		assert.NoError(t, err)

		// Test invalid DID action
		invalidAtt := ucan.CreateDIDAttenuation([]string{"invalid-action"}, "*", nil)
		err = template.ValidateAttenuation(invalidAtt)
		assert.Error(t, err)
	})

	t.Run("TestDWNTemplate", func(t *testing.T) {
		template := ucan.StandardDWNTemplate()

		// Test valid DWN actions
		dwnAtt := ucan.CreateDWNAttenuation([]string{"create", "read"}, "*", nil)
		err := template.ValidateAttenuation(dwnAtt)
		assert.NoError(t, err)
	})

	t.Run("TestDEXTemplate", func(t *testing.T) {
		template := ucan.StandardDEXTemplate()

		// Test valid DEX actions
		dexAtt := ucan.CreateDEXAttenuation([]string{"swap", "provide-liquidity"}, "*", nil, "")
		err := template.ValidateAttenuation(dexAtt)
		assert.NoError(t, err)
	})
}

func TestEnhancedVerification(t *testing.T) {
	t.Run("TestVerifierWithModuleCapabilities", func(t *testing.T) {
		// Create a test token
		attenuations := []ucan.Attenuation{
			ucan.CreateDIDAttenuation([]string{"create"}, "*", []string{"owner"}),
		}

		tokenString, err := ucan.GenerateModuleJWTToken(
			attenuations,
			"did:key:alice",
			"did:key:bob",
			time.Hour,
		)
		require.NoError(t, err)

		// Verify capability - this would normally require proper DID resolution
		// For now, just test that the parsing works
		token, err := ucan.VerifyModuleJWTToken(tokenString, "did:key:alice", "did:key:bob")
		require.NoError(t, err)
		assert.Len(t, token.Attenuations, 1)

		didCap, ok := token.Attenuations[0].Capability.(*ucan.DIDCapability)
		assert.True(t, ok)
		assert.Contains(t, didCap.Actions, "create")
		assert.Contains(t, didCap.Caveats, "owner")
	})
}
