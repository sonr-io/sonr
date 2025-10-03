//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

// PaymentTokenizer handles secure payment method tokenization
type PaymentTokenizer struct {
	mu         sync.RWMutex
	tokens     map[string]*TokenData
	encryptKey []byte
}

// TokenData stores tokenized payment data
type TokenData struct {
	Token       string    `json:"token"`
	LastFour    string    `json:"last_four"`
	Brand       string    `json:"brand"`
	ExpiryMonth int       `json:"expiry_month"`
	ExpiryYear  int       `json:"expiry_year"`
	CreatedAt   time.Time `json:"created_at"`
	UsedCount   int       `json:"used_count"`
}

// TransactionSigner handles transaction signing and verification
type TransactionSigner struct {
	signKey []byte
}

// PCICompliance handles PCI DSS compliance requirements
type PCICompliance struct {
	auditLog []AuditEntry
	mu       sync.RWMutex
}

// AuditEntry for PCI compliance logging
type AuditEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Action     string    `json:"action"`
	UserID     string    `json:"user_id"`
	ResourceID string    `json:"resource_id"`
	Result     string    `json:"result"`
	IPAddress  string    `json:"ip_address"`
}

var (
	paymentTokenizer  *PaymentTokenizer
	transactionSigner *TransactionSigner
	pciCompliance     *PCICompliance
	initOnce          sync.Once
)

// InitializePaymentSecurity initializes payment security components
func InitializePaymentSecurity() {
	initOnce.Do(func() {
		// Generate encryption key (in production, use KMS)
		encKey := make([]byte, 32)
		io.ReadFull(rand.Reader, encKey)

		// Generate signing key
		signKey := make([]byte, 32)
		io.ReadFull(rand.Reader, signKey)

		paymentTokenizer = &PaymentTokenizer{
			tokens:     make(map[string]*TokenData),
			encryptKey: encKey,
		}

		transactionSigner = &TransactionSigner{
			signKey: signKey,
		}

		pciCompliance = &PCICompliance{
			auditLog: make([]AuditEntry, 0),
		}
	})
}

// TokenizeCard tokenizes credit card data (PCI DSS compliant)
func TokenizeCard(cardNumber, cvv string, expiryMonth, expiryYear int) (string, error) {
	// Validate card number using Luhn algorithm
	if !validateLuhn(cardNumber) {
		return "", fmt.Errorf("invalid card number")
	}

	// Validate CVV
	if !validateCVV(cvv) {
		return "", fmt.Errorf("invalid CVV")
	}

	// Validate expiry
	if !validateExpiry(expiryMonth, expiryYear) {
		return "", fmt.Errorf("card expired or invalid expiry date")
	}

	// Extract card info
	lastFour := cardNumber[len(cardNumber)-4:]
	brand := detectCardBrand(cardNumber)

	// Generate secure token
	token := generateSecureToken()

	// Store tokenized data (never store raw card data)
	tokenData := &TokenData{
		Token:       token,
		LastFour:    lastFour,
		Brand:       brand,
		ExpiryMonth: expiryMonth,
		ExpiryYear:  expiryYear,
		CreatedAt:   time.Now(),
		UsedCount:   0,
	}

	paymentTokenizer.mu.Lock()
	paymentTokenizer.tokens[token] = tokenData
	paymentTokenizer.mu.Unlock()

	// Log tokenization for PCI compliance
	pciCompliance.LogAction("TOKENIZE_CARD", "", token, "SUCCESS", "")

	return token, nil
}

// validateLuhn validates credit card number using Luhn algorithm
func validateLuhn(cardNumber string) bool {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// Check if all digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cardNumber) {
		return false
	}

	// Luhn algorithm
	sum := 0
	isEven := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

// validateCVV validates CVV format
func validateCVV(cvv string) bool {
	// CVV should be 3 or 4 digits
	return regexp.MustCompile(`^\d{3,4}$`).MatchString(cvv)
}

// validateExpiry validates card expiry date
func validateExpiry(month, year int) bool {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Check valid month
	if month < 1 || month > 12 {
		return false
	}

	// Check if expired
	if year < currentYear || (year == currentYear && month < currentMonth) {
		return false
	}

	// Check reasonable future date (max 20 years)
	if year > currentYear+20 {
		return false
	}

	return true
}

// detectCardBrand detects card brand from number
func detectCardBrand(cardNumber string) string {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// Visa
	if strings.HasPrefix(cardNumber, "4") {
		return "visa"
	}

	// Mastercard
	if regexp.MustCompile(`^5[1-5]`).MatchString(cardNumber) ||
		regexp.MustCompile(`^2[2-7]`).MatchString(cardNumber) {
		return "mastercard"
	}

	// American Express
	if strings.HasPrefix(cardNumber, "34") || strings.HasPrefix(cardNumber, "37") {
		return "amex"
	}

	// Discover
	if strings.HasPrefix(cardNumber, "6011") || strings.HasPrefix(cardNumber, "65") {
		return "discover"
	}

	return "unknown"
}

// generateSecureToken generates a cryptographically secure token
func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "tok_" + base64.URLEncoding.EncodeToString(b)
}

// SignTransaction signs a transaction for integrity
func SignTransaction(transactionData map[string]interface{}) (string, error) {
	// Serialize transaction data
	data, err := json.Marshal(transactionData)
	if err != nil {
		return "", err
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, transactionSigner.signKey)
	h.Write(data)
	signature := hex.EncodeToString(h.Sum(nil))

	// Log signing for audit
	txID := ""
	if id, ok := transactionData["id"].(string); ok {
		txID = id
	}
	pciCompliance.LogAction("SIGN_TRANSACTION", "", txID, "SUCCESS", "")

	return signature, nil
}

// VerifyTransactionSignature verifies a transaction signature
func VerifyTransactionSignature(transactionData map[string]interface{}, signature string) bool {
	// Serialize transaction data
	data, err := json.Marshal(transactionData)
	if err != nil {
		return false
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, transactionSigner.signKey)
	h.Write(data)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// EncryptSensitiveData encrypts sensitive payment data
func EncryptSensitiveData(plaintext string) (string, error) {
	// Create cipher
	block, err := aes.NewCipher(paymentTokenizer.encryptKey)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Combine nonce and ciphertext
	combined := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptSensitiveData decrypts sensitive payment data
func DecryptSensitiveData(encrypted string) (string, error) {
	// Decode from base64
	combined, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	// Extract nonce and ciphertext
	if len(combined) < 12 {
		return "", fmt.Errorf("invalid encrypted data")
	}

	nonce := combined[:12]
	ciphertext := combined[12:]

	// Create cipher
	block, err := aes.NewCipher(paymentTokenizer.encryptKey)
	if err != nil {
		return "", err
	}

	// Decrypt
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// LogAction logs an action for PCI compliance audit
func (p *PCICompliance) LogAction(action, userID, resourceID, result, ipAddress string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry := AuditEntry{
		Timestamp:  time.Now(),
		Action:     action,
		UserID:     userID,
		ResourceID: resourceID,
		Result:     result,
		IPAddress:  ipAddress,
	}

	p.auditLog = append(p.auditLog, entry)

	// In production, persist to secure audit log storage
	// For now, just keep in memory (limited to last 10000 entries)
	if len(p.auditLog) > 10000 {
		p.auditLog = p.auditLog[1:]
	}
}

// GetAuditLog returns recent audit log entries
func (p *PCICompliance) GetAuditLog(limit int) []AuditEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit > len(p.auditLog) {
		limit = len(p.auditLog)
	}

	// Return most recent entries
	start := len(p.auditLog) - limit
	if start < 0 {
		start = 0
	}

	return p.auditLog[start:]
}

// ValidateToken validates a payment token
func ValidateToken(token string) (*TokenData, error) {
	paymentTokenizer.mu.RLock()
	defer paymentTokenizer.mu.RUnlock()

	tokenData, exists := paymentTokenizer.tokens[token]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is expired (tokens valid for 1 hour)
	if time.Since(tokenData.CreatedAt) > time.Hour {
		return nil, fmt.Errorf("token expired")
	}

	// Increment usage count
	tokenData.UsedCount++

	return tokenData, nil
}

// MaskCardNumber masks all but last 4 digits of card number
func MaskCardNumber(cardNumber string) string {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	if len(cardNumber) < 4 {
		return strings.Repeat("*", len(cardNumber))
	}

	lastFour := cardNumber[len(cardNumber)-4:]
	masked := strings.Repeat("*", len(cardNumber)-4) + lastFour

	// Format based on card type
	if len(masked) == 16 {
		// Format as XXXX XXXX XXXX 1234
		return masked[:4] + " " + masked[4:8] + " " + masked[8:12] + " " + masked[12:]
	}

	return masked
}

// SanitizePaymentData removes sensitive data from payment objects
func SanitizePaymentData(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	// List of sensitive fields to exclude
	sensitiveFields := []string{
		"card_number", "cvv", "cvc", "card_code",
		"account_number", "routing_number", "pin",
	}

	for key, value := range data {
		// Check if field is sensitive
		isSensitive := false
		keyLower := strings.ToLower(key)
		for _, sensitive := range sensitiveFields {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true
				break
			}
		}

		if !isSensitive {
			sanitized[key] = value
		}
	}

	return sanitized
}
