//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*RequestCounter
	limit    int
	window   time.Duration
}

// RequestCounter tracks requests
type RequestCounter struct {
	Count     int
	ResetTime time.Time
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EnableRateLimit bool
	RateLimit       int
	RateWindow      time.Duration
	EnableCSP       bool
	CSPPolicy       string
}

// Global security configuration
var securityConfig = &SecurityConfig{
	EnableRateLimit: true,
	RateLimit:       100, // 100 requests
	RateWindow:      time.Minute,
	EnableCSP:       true,
	CSPPolicy:       "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; connect-src 'self' https:; img-src 'self' data: https:; style-src 'self' 'unsafe-inline';",
}

// Global rate limiter
var rateLimiter = NewRateLimiter(securityConfig.RateLimit, securityConfig.RateWindow)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*RequestCounter),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if request is allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	counter, exists := rl.requests[identifier]

	if !exists || now.After(counter.ResetTime) {
		// Create new counter or reset existing one
		rl.requests[identifier] = &RequestCounter{
			Count:     1,
			ResetTime: now.Add(rl.window),
		}
		return true
	}

	if counter.Count >= rl.limit {
		return false
	}

	counter.Count++
	return true
}

// SecurityMiddleware wraps handlers with security features
func SecurityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply security headers
		applySecurityHeaders(w)

		// Rate limiting
		if securityConfig.EnableRateLimit {
			// Use client IP or a default identifier for WASM environment
			identifier := getClientIdentifier(r)
			if !rateLimiter.Allow(identifier) {
				writeError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}
		}

		// Call the next handler
		next(w, r)
	}
}

// applySecurityHeaders applies security headers to response
func applySecurityHeaders(w http.ResponseWriter) {
	// CORS headers (already handled by handleCORS, but adding for completeness)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	// Content Security Policy
	if securityConfig.EnableCSP {
		w.Header().Set("Content-Security-Policy", securityConfig.CSPPolicy)
	}

	// Strict Transport Security (for HTTPS)
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
}

// getClientIdentifier gets a client identifier for rate limiting
func getClientIdentifier(r *http.Request) string {
	// In WASM environment, we can't rely on real IP
	// Use a combination of headers for identification

	// Try X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Try X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Try Origin header (common in browser requests)
	if origin := r.Header.Get("Origin"); origin != "" {
		return origin
	}

	// Try User-Agent as last resort
	if ua := r.Header.Get("User-Agent"); ua != "" {
		return ua
	}

	// Default identifier
	return "default-client"
}

// ValidatePaymentData validates payment data for security
func ValidatePaymentData(data map[string]interface{}) error {
	// Check for required fields
	requiredFields := []string{"amount", "currency"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate amount
	if amount, ok := data["amount"].(float64); ok {
		if amount <= 0 || amount > 1000000 {
			return fmt.Errorf("invalid amount")
		}
	}

	// Validate currency
	if currency, ok := data["currency"].(string); ok {
		validCurrencies := []string{"USD", "EUR", "GBP", "JPY"}
		valid := false
		for _, vc := range validCurrencies {
			if currency == vc {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported currency")
		}
	}

	return nil
}

// SanitizeInput sanitizes user input
func SanitizeInput(input string) string {
	// Remove any potentially dangerous characters
	// This is a basic implementation - in production, use a proper sanitization library
	sanitized := input

	// Remove script tags
	sanitized = strings.ReplaceAll(sanitized, "<script>", "")
	sanitized = strings.ReplaceAll(sanitized, "</script>", "")

	// Remove other potentially dangerous HTML
	sanitized = strings.ReplaceAll(sanitized, "<iframe>", "")
	sanitized = strings.ReplaceAll(sanitized, "</iframe>", "")

	// Limit length
	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000]
	}

	return sanitized
}

// TokenizePaymentMethod creates a token for payment method
func TokenizePaymentMethod(method map[string]interface{}) string {
	// Create a secure token representing the payment method
	// In production, this would use proper tokenization service

	token := generateRandomString(32)

	// Store token mapping (in production, use secure storage)
	// For now, just return the token
	return "pmtoken_" + token
}

// ValidateOrigin validates request origin
func ValidateOrigin(origin string) bool {
	// List of allowed origins
	allowedOrigins := []string{
		"https://motor.sonr.io",
		"https://localhost:3000",
		"http://localhost:3000",
		"https://sonr.io",
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}
