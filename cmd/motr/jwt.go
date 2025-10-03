//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	kid        string
	issuer     string
}

// JWTHeader represents JWT header
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid,omitempty"`
}

// JWTClaims represents standard JWT claims
type JWTClaims struct {
	Issuer     string                 `json:"iss,omitempty"`
	Subject    string                 `json:"sub,omitempty"`
	Audience   interface{}            `json:"aud,omitempty"` // Can be string or []string
	Expiration int64                  `json:"exp,omitempty"`
	NotBefore  int64                  `json:"nbf,omitempty"`
	IssuedAt   int64                  `json:"iat,omitempty"`
	JWTID      string                 `json:"jti,omitempty"`
	Nonce      string                 `json:"nonce,omitempty"`
	Extra      map[string]interface{} `json:"-"`
}

// IDToken represents an OpenID Connect ID token
type IDToken struct {
	JWTClaims
	AuthTime      int64    `json:"auth_time,omitempty"`
	Nonce         string   `json:"nonce,omitempty"`
	ACR           string   `json:"acr,omitempty"`
	AMR           []string `json:"amr,omitempty"`
	AZP           string   `json:"azp,omitempty"`
	Name          string   `json:"name,omitempty"`
	GivenName     string   `json:"given_name,omitempty"`
	FamilyName    string   `json:"family_name,omitempty"`
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
}

// Global JWT manager instance
var jwtManager *JWTManager

// InitJWTManager initializes the JWT manager
func InitJWTManager() error {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	jwtManager = &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		kid:        "motor-key-1",
		issuer:     "https://motor.sonr.io",
	}

	return nil
}

// GenerateToken generates a JWT token
func (m *JWTManager) GenerateToken(claims JWTClaims) (string, error) {
	// Set standard claims
	if claims.Issuer == "" {
		claims.Issuer = m.issuer
	}
	if claims.IssuedAt == 0 {
		claims.IssuedAt = time.Now().Unix()
	}
	if claims.Expiration == 0 {
		claims.Expiration = time.Now().Add(1 * time.Hour).Unix()
	}

	// Create header
	header := JWTHeader{
		Alg: "RS256",
		Typ: "JWT",
		Kid: m.kid,
	}

	// Encode header
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, _ := json.Marshal(claims)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, m.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	// Combine parts
	token := message + "." + signatureEncoded
	return token, nil
}

// GenerateIDToken generates an OpenID Connect ID token
func (m *JWTManager) GenerateIDToken(subject, audience, nonce string, extra map[string]interface{}) (string, error) {
	idToken := IDToken{
		JWTClaims: JWTClaims{
			Issuer:     m.issuer,
			Subject:    subject,
			Audience:   audience,
			IssuedAt:   time.Now().Unix(),
			Expiration: time.Now().Add(1 * time.Hour).Unix(),
			Nonce:      nonce,
		},
		AuthTime:      time.Now().Unix(),
		Email:         fmt.Sprintf("%s@motor.sonr.io", subject),
		EmailVerified: true,
	}

	// Convert to claims
	claims := JWTClaims{
		Issuer:     idToken.Issuer,
		Subject:    idToken.Subject,
		Audience:   idToken.Audience,
		IssuedAt:   idToken.IssuedAt,
		Expiration: idToken.Expiration,
		Nonce:      idToken.Nonce,
		Extra: map[string]interface{}{
			"auth_time":      idToken.AuthTime,
			"email":          idToken.Email,
			"email_verified": idToken.EmailVerified,
		},
	}

	// Add extra claims
	for k, v := range extra {
		claims.Extra[k] = v
	}

	return m.GenerateToken(claims)
}

// ValidateToken validates a JWT token
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Split token
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Verify algorithm
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Verify signature
	message := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	hash := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(m.publicKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	// Verify expiration
	if claims.Expiration > 0 && time.Now().Unix() > claims.Expiration {
		return nil, fmt.Errorf("token expired")
	}

	// Verify not before
	if claims.NotBefore > 0 && time.Now().Unix() < claims.NotBefore {
		return nil, fmt.Errorf("token not yet valid")
	}

	return &claims, nil
}

// GetPublicKeyJWK returns the public key in JWK format
func (m *JWTManager) GetPublicKeyJWK() map[string]interface{} {
	// Get modulus and exponent
	n := base64.RawURLEncoding.EncodeToString(m.publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}) // 65537

	return map[string]interface{}{
		"kty": "RSA",
		"use": "sig",
		"kid": m.kid,
		"alg": "RS256",
		"n":   n,
		"e":   e,
	}
}

// GetPublicKeyPEM returns the public key in PEM format
func (m *JWTManager) GetPublicKeyPEM() string {
	pubKeyBytes, _ := x509.MarshalPKIXPublicKey(m.publicKey)
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	return string(pubKeyPEM)
}

// GenerateAccessToken generates an access token
func (m *JWTManager) GenerateAccessToken(subject, scope string) (string, error) {
	claims := JWTClaims{
		Subject: subject,
		Extra: map[string]interface{}{
			"scope":      scope,
			"token_type": "Bearer",
		},
	}
	return m.GenerateToken(claims)
}

// GenerateRefreshToken generates a refresh token
func (m *JWTManager) GenerateRefreshToken(subject string) (string, error) {
	claims := JWTClaims{
		Subject:    subject,
		Expiration: time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 days
		Extra: map[string]interface{}{
			"token_type": "refresh",
		},
	}
	return m.GenerateToken(claims)
}
