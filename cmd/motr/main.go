//go:build js && wasm
// +build js,wasm

package main

import (
	"log"
	"net/http"

	wasmhttp "github.com/go-sonr/wasm-http-server/v3"
)

func main() {
	// Set up HTTP routes
	setupRoutes()

	// Start the WASM HTTP server
	log.Println("Motor Payment Gateway & OIDC Server starting...")
	log.Println("Available endpoints:")
	log.Println("  Health: /health, /status")
	log.Println("  Payment API: /api/payment/*")
	log.Println("  OIDC: /.well-known/*, /authorize, /token, /userinfo")

	wasmhttp.Serve(nil)
}

// setupRoutes configures all HTTP routes with security middleware
func setupRoutes() {
	// Health and status endpoints (no rate limiting)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/status", handleStatus)

	// W3C Payment Handler API endpoints with security
	http.HandleFunc("/payment/instruments", SecurityMiddleware(handlePaymentInstruments))
	http.HandleFunc("/payment/canmakepayment", SecurityMiddleware(handleCanMakePayment))
	http.HandleFunc("/payment/paymentrequest", SecurityMiddleware(handlePaymentRequest))

	// Payment Gateway endpoints with security
	http.HandleFunc("/api/payment/process", SecurityMiddleware(handlePaymentProcess))
	http.HandleFunc("/api/payment/validate", SecurityMiddleware(handlePaymentValidate))
	http.HandleFunc("/api/payment/status/", SecurityMiddleware(handlePaymentStatus))
	http.HandleFunc("/api/payment/refund", SecurityMiddleware(handlePaymentRefund))

	// OIDC endpoints with security
	http.HandleFunc("/.well-known/openid-configuration", handleOIDCDiscovery) // No rate limit for discovery
	http.HandleFunc("/.well-known/jwks.json", handleJWKS)                     // No rate limit for JWKS
	http.HandleFunc("/authorize", SecurityMiddleware(handleAuthorize))
	http.HandleFunc("/token", SecurityMiddleware(handleToken))
	http.HandleFunc("/userinfo", SecurityMiddleware(handleUserInfo))
}
