package auth

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebAuthnClient_BasicRegistration(t *testing.T) {
	client := &webAuthnClient{
		origin: "http://localhost",
		rpID:   "localhost",
		rpName: "Sonr",
	}

	// Test BeginRegistration
	opts := &RegistrationOptions{
		UserID:      "test_user",
		Username:    "testuser",
		DisplayName: "Test User",
	}

	challenge, err := client.BeginRegistration(context.Background(), opts)
	require.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.NotEmpty(t, challenge.Challenge)
	assert.Equal(t, "localhost", challenge.RelyingParty.ID)
	assert.Equal(t, "Sonr", challenge.RelyingParty.Name)
	assert.Equal(t, []byte("test_user"), challenge.User.ID)
	assert.Equal(t, "testuser", challenge.User.Name)
	assert.Equal(t, "Test User", challenge.User.DisplayName)
}

func TestWebAuthnClient_BasicAuthentication(t *testing.T) {
	client := &webAuthnClient{
		origin: "http://localhost",
		rpID:   "localhost",
	}

	// Test BeginAuthentication
	opts := &AuthenticationOptions{
		UserVerification: "preferred",
		Timeout:          60000,
	}

	challenge, err := client.BeginAuthentication(context.Background(), opts)
	require.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.NotEmpty(t, challenge.Challenge)
	assert.Equal(t, "localhost", challenge.RelyingPartyID)
	assert.Equal(t, "preferred", challenge.UserVerification)
	assert.Equal(t, 60000, challenge.Timeout)
}

func TestWebAuthnClient_VerifyClientData(t *testing.T) {
	challenge := []byte("test_challenge")
	origin := "http://localhost"

	// Create valid client data
	clientData := map[string]any{
		"type":      "webauthn.create",
		"challenge": "dGVzdF9jaGFsbGVuZ2U", // base64url encoded "test_challenge"
		"origin":    origin,
	}
	clientDataJSON, _ := json.Marshal(clientData)

	// Test successful verification
	result, err := verifyClientData(clientDataJSON, challenge, "webauthn.create", origin)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "webauthn.create", string(result.Type))
	assert.Equal(t, origin, result.Origin)

	// Test wrong type
	clientData["type"] = "webauthn.get"
	clientDataJSON, _ = json.Marshal(clientData)
	_, err = verifyClientData(clientDataJSON, challenge, "webauthn.create", origin)
	assert.Error(t, err)

	// Test wrong origin
	clientData["type"] = "webauthn.create"
	clientData["origin"] = "http://evil.com"
	clientDataJSON, _ = json.Marshal(clientData)
	_, err = verifyClientData(clientDataJSON, challenge, "webauthn.create", origin)
	assert.Error(t, err)

	// Test wrong challenge
	clientData["origin"] = origin
	clientData["challenge"] = "d3JvbmdfY2hhbGxlbmdl" // base64url encoded "wrong_challenge"
	clientDataJSON, _ = json.Marshal(clientData)
	_, err = verifyClientData(clientDataJSON, challenge, "webauthn.create", origin)
	assert.Error(t, err)
}
