//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %v", response["status"])
	}
}

// TestPaymentInstruments tests getting payment instruments
func TestPaymentInstruments(t *testing.T) {
	req := httptest.NewRequest("GET", "/payment/instruments", nil)
	w := httptest.NewRecorder()

	handlePaymentInstruments(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var instruments []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&instruments); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(instruments) == 0 {
		t.Error("Expected at least one payment instrument")
	}
}

// TestCanMakePayment tests payment capability check
func TestCanMakePayment(t *testing.T) {
	payload := map[string]interface{}{
		"origin": "https://localhost:3000",
		"methodData": []map[string]interface{}{
			{
				"supportedMethods": "https://motor.sonr.io/pay",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/payment/canmakepayment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleCanMakePayment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["canMakePayment"] != true {
		t.Errorf("Expected canMakePayment to be true")
	}
}

// TestProcessPayment tests payment processing with security
func TestProcessPayment(t *testing.T) {
	// Initialize payment security
	InitializePaymentSecurity()

	payload := map[string]interface{}{
		"origin":           "https://localhost:3000",
		"topOrigin":        "https://localhost:3000",
		"paymentRequestId": "test-request-123",
		"methodData": []map[string]interface{}{
			{
				"supportedMethods": "https://motor.sonr.io/pay",
			},
		},
		"details": map[string]interface{}{
			"total": map[string]interface{}{
				"label": "Test Payment",
				"amount": map[string]interface{}{
					"currency": "USD",
					"value":    "100.00",
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/payment/process", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://localhost:3000")
	w := httptest.NewRecorder()

	handleProcessPayment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["paymentId"] == "" {
		t.Error("Expected payment ID in response")
	}

	if response["status"] != "pending" {
		t.Errorf("Expected status pending, got %v", response["status"])
	}
}

// TestCardTokenization tests PCI-compliant card tokenization
func TestCardTokenization(t *testing.T) {
	// Initialize payment security
	InitializePaymentSecurity()

	// Test valid card
	token, err := TokenizeCard("4111111111111111", "123", 12, 2025)
	if err != nil {
		t.Fatalf("Failed to tokenize valid card: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}

	// Test invalid card number
	_, err = TokenizeCard("1234567890123456", "123", 12, 2025)
	if err == nil {
		t.Error("Expected error for invalid card number")
	}

	// Test expired card
	_, err = TokenizeCard("4111111111111111", "123", 1, 2020)
	if err == nil {
		t.Error("Expected error for expired card")
	}
}

// TestTransactionSigning tests transaction signature verification
func TestTransactionSigning(t *testing.T) {
	// Initialize payment security
	InitializePaymentSecurity()

	txData := map[string]interface{}{
		"id":        "test-tx-123",
		"amount":    "100.00",
		"currency":  "USD",
		"method":    "card",
		"timestamp": time.Now().Unix(),
	}

	// Sign transaction
	signature, err := SignTransaction(txData)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	if signature == "" {
		t.Error("Expected signature to be generated")
	}

	// Verify signature
	valid := VerifyTransactionSignature(txData, signature)
	if !valid {
		t.Error("Expected signature to be valid")
	}

	// Test invalid signature
	invalid := VerifyTransactionSignature(txData, "invalid-signature")
	if invalid {
		t.Error("Expected invalid signature to fail verification")
	}
}

// TestOIDCDiscovery tests OIDC discovery endpoint
func TestOIDCDiscovery(t *testing.T) {
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()

	handleOIDCDiscovery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var config map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&config); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check required OIDC fields
	requiredFields := []string{
		"issuer",
		"authorization_endpoint",
		"token_endpoint",
		"userinfo_endpoint",
		"jwks_uri",
	}

	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			t.Errorf("Missing required OIDC field: %s", field)
		}
	}
}

// TestJWKS tests JWKS endpoint
func TestJWKS(t *testing.T) {
	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	handleJWKS(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var jwks map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&jwks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	keys, ok := jwks["keys"].([]interface{})
	if !ok || len(keys) == 0 {
		t.Error("Expected at least one key in JWKS")
	}
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	// Initialize with low rate limit for testing
	securityConfig.RateLimit = 5
	securityConfig.RateWindow = time.Second
	rateLimiter = NewRateLimiter(5, time.Second)

	// Make requests up to the limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "test-client")
		w := httptest.NewRecorder()

		SecurityMiddleware(handleHealth)(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, w.Code)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "test-client")
	w := httptest.NewRecorder()

	SecurityMiddleware(handleHealth)(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected rate limit (429), got %d", w.Code)
	}

	// Wait for rate limit window to reset
	time.Sleep(time.Second + 100*time.Millisecond)

	// Should work again
	req = httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "test-client")
	w = httptest.NewRecorder()

	SecurityMiddleware(handleHealth)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("After reset: Expected status 200, got %d", w.Code)
	}
}

// TestSecurityHeaders tests security headers are properly set
func TestSecurityHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	SecurityMiddleware(handleHealth)(w, req)

	// Check security headers
	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expected := range headers {
		actual := w.Header().Get(header)
		if actual != expected {
			t.Errorf("Header %s: expected %s, got %s", header, expected, actual)
		}
	}

	// Check CSP is present
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("Expected Content-Security-Policy header")
	}
}

// TestPCICompliance tests PCI compliance audit logging
func TestPCICompliance(t *testing.T) {
	// Initialize payment security
	InitializePaymentSecurity()

	// Log some actions
	pciCompliance.LogAction("TEST_ACTION", "user123", "resource456", "SUCCESS", "127.0.0.1")

	// Get audit log
	logs := pciCompliance.GetAuditLog(10)

	if len(logs) == 0 {
		t.Error("Expected audit log entries")
	}

	// Check last entry
	lastLog := logs[len(logs)-1]
	if lastLog.Action != "TEST_ACTION" {
		t.Errorf("Expected action TEST_ACTION, got %s", lastLog.Action)
	}

	if lastLog.UserID != "user123" {
		t.Errorf("Expected user ID user123, got %s", lastLog.UserID)
	}
}

// TestDataEncryption tests sensitive data encryption
func TestDataEncryption(t *testing.T) {
	// Initialize payment security
	InitializePaymentSecurity()

	sensitiveData := "4111-1111-1111-1111"

	// Encrypt data
	encrypted, err := EncryptSensitiveData(sensitiveData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	if encrypted == sensitiveData {
		t.Error("Encrypted data should not match plaintext")
	}

	// Decrypt data
	decrypted, err := DecryptSensitiveData(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	if decrypted != sensitiveData {
		t.Errorf("Decrypted data doesn't match original: got %s, want %s", decrypted, sensitiveData)
	}
}

// TestCardMasking tests card number masking
func TestCardMasking(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"4111111111111111", "**** **** **** 1111"},
		{"5500000000000004", "**** **** **** 0004"},
		{"340000000000009", "********** 00009"},
		{"123", "123"}, // Too short to mask
	}

	for _, tc := range testCases {
		masked := MaskCardNumber(tc.input)
		if masked != tc.expected {
			t.Errorf("MaskCardNumber(%s): got %s, want %s", tc.input, masked, tc.expected)
		}
	}
}
