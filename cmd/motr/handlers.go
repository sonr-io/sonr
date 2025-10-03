//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Health & Status Handlers

// handleHealth returns service health status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "motor-gateway",
		"timestamp": time.Now().Unix(),
	})
}

// handleStatus returns detailed service status
func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "operational",
		"version": "1.0.0",
		"services": map[string]string{
			"payment_gateway": "active",
			"oidc_provider":   "active",
		},
		"uptime": time.Now().Unix(),
	})
}

// W3C Payment Handler API Handlers

// handlePaymentInstruments returns available payment instruments
func handlePaymentInstruments(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	instruments := paymentHandler.GetInstruments()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"instruments": instruments,
	})
}

// handleCanMakePayment checks if payment can be made
func handleCanMakePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		MethodData []PaymentMethod `json:"methodData"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	canMakePayment := paymentHandler.CanMakePayment(req.MethodData)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"canMakePayment": canMakePayment,
	})
}

// handlePaymentRequest handles W3C PaymentRequestEvent
func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse payment request event
	var reqData json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	paymentReq, err := SerializePaymentRequest(reqData)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payment request")
		return
	}

	// Process payment request
	tx, err := paymentHandler.ProcessPayment(paymentReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Payment processing failed")
		return
	}

	// Return payment response
	if tx.Response != nil {
		writeJSON(w, http.StatusOK, tx.Response)
	} else {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"transactionId": tx.ID,
			"status":        tx.Status,
		})
	}
}

// Payment Gateway Handlers

// handlePaymentProcess processes a payment transaction using W3C Payment Handler API
func handlePaymentProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse payment request
	var reqData json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	paymentReq, err := SerializePaymentRequest(reqData)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payment request")
		return
	}

	// Process payment
	tx, err := paymentHandler.ProcessPayment(paymentReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Payment processing failed")
		return
	}

	writeJSON(w, http.StatusOK, tx)
}

// handlePaymentValidate validates a payment method
func handlePaymentValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse validation request
	var req struct {
		Method string                 `json:"method"`
		Data   map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate payment method
	valid, err := paymentHandler.ValidatePaymentMethod(req.Method, req.Data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Validation failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":   valid,
		"method":  req.Method,
		"message": "Payment method validation complete",
	})
}

// handlePaymentStatus returns payment transaction status
func handlePaymentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract transaction ID from path
	txID := strings.TrimPrefix(r.URL.Path, "/api/payment/status/")
	if txID == "" {
		writeError(w, http.StatusBadRequest, "Transaction ID required")
		return
	}

	// Get transaction from handler
	tx, exists := paymentHandler.GetTransaction(txID)
	if !exists {
		writeError(w, http.StatusNotFound, "Transaction not found")
		return
	}

	writeJSON(w, http.StatusOK, tx)
}

// handlePaymentRefund processes a refund
func handlePaymentRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: Implement refund processing
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"refund_id": "ref_" + generateID(),
		"status":    "processing",
		"message":   "Refund initiated",
	})
}

// OIDC Handlers

// handleOIDCDiscovery returns OIDC discovery document
func handleOIDCDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	discovery := oidcProvider.GetDiscovery()
	writeJSON(w, http.StatusOK, discovery)
}

// handleJWKS returns JSON Web Key Set
func handleJWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	jwk := jwtManager.GetPublicKeyJWK()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"keys": []map[string]interface{}{jwk},
	})
}

// handleAuthorize handles authorization requests
func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "GET" && r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse authorization request
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	responseType := r.FormValue("response_type")
	scope := r.FormValue("scope")
	state := r.FormValue("state")
	nonce := r.FormValue("nonce")
	codeChallenge := r.FormValue("code_challenge")
	codeChallengeMethod := r.FormValue("code_challenge_method")

	// Validate request
	if clientID == "" || redirectURI == "" || responseType == "" {
		writeError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	// For demo, auto-approve with test user
	userID := "test-user"

	// Generate authorization code
	authCode, err := oidcProvider.GenerateAuthorizationCode(
		clientID, redirectURI, scope, state, nonce, userID,
		codeChallenge, codeChallengeMethod,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return authorization code
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":         authCode.Code,
		"state":        state,
		"redirect_uri": redirectURI,
	})
}

// handleToken handles token requests
func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse token request
	var req TokenRequest
	req.GrantType = r.FormValue("grant_type")
	req.Code = r.FormValue("code")
	req.RedirectURI = r.FormValue("redirect_uri")
	req.ClientID = r.FormValue("client_id")
	req.ClientSecret = r.FormValue("client_secret")
	req.RefreshToken = r.FormValue("refresh_token")
	req.Scope = r.FormValue("scope")
	req.CodeVerifier = r.FormValue("code_verifier")

	// Handle based on grant type
	var resp *TokenResponse
	var err error

	switch req.GrantType {
	case "authorization_code":
		resp, err = oidcProvider.ExchangeCode(&req)
	case "refresh_token":
		// TODO: Implement refresh token flow
		writeError(w, http.StatusNotImplemented, "Refresh token not yet implemented")
		return
	default:
		writeError(w, http.StatusBadRequest, "Unsupported grant type")
		return
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleUserInfo returns user information
func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		handleCORS(w)
		return
	}

	if r.Method != "GET" && r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get bearer token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "Missing authorization header")
		return
	}

	// Extract token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeError(w, http.StatusUnauthorized, "Invalid authorization header")
		return
	}

	accessToken := parts[1]

	// Get user info
	userInfo, err := oidcProvider.GetUserInfo(accessToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, userInfo)
}

// Helper Functions

// handleCORS handles CORS preflight requests
func handleCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// generateID generates a simple ID
func generateID() string {
	return time.Now().Format("20060102150405")
}
