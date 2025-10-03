//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PaymentMethod represents a payment method according to W3C Payment Handler API
type PaymentMethod struct {
	SupportedMethods string      `json:"supportedMethods"`
	Data             interface{} `json:"data,omitempty"`
}

// PaymentDetails contains payment details
type PaymentDetails struct {
	Total           PaymentItem   `json:"total"`
	DisplayItems    []PaymentItem `json:"displayItems,omitempty"`
	Modifiers       []interface{} `json:"modifiers,omitempty"`
	ShippingOptions []interface{} `json:"shippingOptions,omitempty"`
}

// PaymentItem represents an item in payment
type PaymentItem struct {
	Label  string          `json:"label"`
	Amount PaymentCurrency `json:"amount"`
}

// PaymentCurrency represents currency amount
type PaymentCurrency struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

// PaymentRequest represents a W3C Payment Request
type PaymentRequest struct {
	ID               string          `json:"id"`
	MethodData       []PaymentMethod `json:"methodData"`
	Details          PaymentDetails  `json:"details"`
	Options          PaymentOptions  `json:"options,omitempty"`
	Origin           string          `json:"origin"`
	TopOrigin        string          `json:"topOrigin"`
	PaymentRequestID string          `json:"paymentRequestId"`
	Total            PaymentItem     `json:"total"`
}

// PaymentOptions contains payment options
type PaymentOptions struct {
	RequestPayerName  bool   `json:"requestPayerName,omitempty"`
	RequestPayerEmail bool   `json:"requestPayerEmail,omitempty"`
	RequestPayerPhone bool   `json:"requestPayerPhone,omitempty"`
	RequestShipping   bool   `json:"requestShipping,omitempty"`
	ShippingType      string `json:"shippingType,omitempty"`
}

// PaymentResponse represents response to payment request
type PaymentResponse struct {
	RequestID       string                 `json:"requestId"`
	MethodName      string                 `json:"methodName"`
	Details         map[string]interface{} `json:"details"`
	PayerName       string                 `json:"payerName,omitempty"`
	PayerEmail      string                 `json:"payerEmail,omitempty"`
	PayerPhone      string                 `json:"payerPhone,omitempty"`
	ShippingAddress interface{}            `json:"shippingAddress,omitempty"`
}

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status"`
	Amount    PaymentCurrency        `json:"amount"`
	Method    string                 `json:"method"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Request   *PaymentRequest        `json:"request,omitempty"`
	Response  *PaymentResponse       `json:"response,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentHandler manages payment processing
type PaymentHandler struct {
	mu           sync.RWMutex
	transactions map[string]*PaymentTransaction
	instruments  []PaymentInstrument
}

// PaymentInstrument represents a payment instrument
type PaymentInstrument struct {
	Name         string   `json:"name"`
	Icons        []Icon   `json:"icons,omitempty"`
	Method       string   `json:"method"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// Icon represents a payment instrument icon
type Icon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes,omitempty"`
	Type  string `json:"type,omitempty"`
}

// Global payment handler instance
var paymentHandler = &PaymentHandler{
	transactions: make(map[string]*PaymentTransaction),
	instruments: []PaymentInstrument{
		{
			Name:         "Motor Payment",
			Method:       "https://motor.sonr.io/pay",
			Capabilities: []string{"basic-card", "tokenized-card"},
		},
	},
}

// ProcessPayment processes a payment request with enhanced security
func (h *PaymentHandler) ProcessPayment(req *PaymentRequest) (*PaymentTransaction, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize payment security if not already done
	InitializePaymentSecurity()

	// Validate origin for security
	if !ValidateOrigin(req.Origin) {
		return nil, fmt.Errorf("invalid origin: %s", req.Origin)
	}

	// Generate transaction ID
	txID := generateTransactionID()

	// Create transaction data for signing
	txData := map[string]interface{}{
		"id":        txID,
		"amount":    req.Details.Total.Amount.Value,
		"currency":  req.Details.Total.Amount.Currency,
		"method":    req.MethodData[0].SupportedMethods,
		"timestamp": time.Now().Unix(),
	}

	// Sign transaction for integrity
	signature, err := SignTransaction(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Create transaction
	tx := &PaymentTransaction{
		ID:        txID,
		Status:    "pending",
		Amount:    req.Details.Total.Amount,
		Method:    req.MethodData[0].SupportedMethods,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Request:   req,
		Metadata: map[string]interface{}{
			"origin":    req.Origin,
			"topOrigin": req.TopOrigin,
			"signature": signature,
		},
	}

	// Log for PCI compliance
	pciCompliance.LogAction("PROCESS_PAYMENT", "", txID, "INITIATED", req.Origin)

	// Store transaction
	h.transactions[txID] = tx

	// Process payment asynchronously with security checks
	go h.processPaymentSecurely(txID)

	return tx, nil
}

// ValidatePaymentMethod validates a payment method
func (h *PaymentHandler) ValidatePaymentMethod(method string, data interface{}) (bool, error) {
	// Check if method is supported
	for _, instrument := range h.instruments {
		if instrument.Method == method {
			// Perform validation based on method type
			switch method {
			case "basic-card", "https://motor.sonr.io/pay":
				return h.validateCardData(data)
			default:
				return true, nil
			}
		}
	}
	return false, nil
}

// GetTransaction retrieves a transaction by ID
func (h *PaymentHandler) GetTransaction(id string) (*PaymentTransaction, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tx, exists := h.transactions[id]
	return tx, exists
}

// UpdateTransactionStatus updates transaction status
func (h *PaymentHandler) UpdateTransactionStatus(id, status string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if tx, exists := h.transactions[id]; exists {
		tx.Status = status
		tx.UpdatedAt = time.Now()
		return nil
	}
	return nil
}

// CanMakePayment checks if payment can be made
func (h *PaymentHandler) CanMakePayment(methods []PaymentMethod) bool {
	for _, method := range methods {
		for _, instrument := range h.instruments {
			if instrument.Method == method.SupportedMethods {
				return true
			}
		}
	}
	return false
}

// GetInstruments returns available payment instruments
func (h *PaymentHandler) GetInstruments() []PaymentInstrument {
	return h.instruments
}

// Helper functions

// generateTransactionID generates a unique transaction ID
func generateTransactionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "txn_" + hex.EncodeToString(bytes)
}

// validateCardData validates and tokenizes card payment data
func (h *PaymentHandler) validateCardData(data interface{}) (bool, error) {
	// Initialize payment security
	InitializePaymentSecurity()

	if data == nil {
		return false, fmt.Errorf("no payment data provided")
	}

	// Parse card data
	cardData, ok := data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid payment data format")
	}

	// Extract card details
	cardNumber, hasNumber := cardData["cardNumber"].(string)
	cvv, hasCVV := cardData["cvv"].(string)
	expiryMonth, hasMonth := cardData["expiryMonth"].(float64)
	expiryYear, hasYear := cardData["expiryYear"].(float64)

	if !hasNumber || !hasCVV || !hasMonth || !hasYear {
		return false, fmt.Errorf("missing required card fields")
	}

	// Tokenize the card for PCI compliance
	token, err := TokenizeCard(cardNumber, cvv, int(expiryMonth), int(expiryYear))
	if err != nil {
		return false, fmt.Errorf("card validation failed: %v", err)
	}

	// Replace sensitive data with token
	cardData["token"] = token
	cardData["cardNumber"] = MaskCardNumber(cardNumber)
	delete(cardData, "cvv") // Never store CVV

	return true, nil
}

// processPaymentSecurely processes payment with enhanced security
func (h *PaymentHandler) processPaymentSecurely(txID string) {
	// Initialize payment security
	InitializePaymentSecurity()

	// Simulate processing delay
	time.Sleep(2 * time.Second)

	// Verify transaction exists
	h.mu.RLock()
	tx, exists := h.transactions[txID]
	h.mu.RUnlock()

	if !exists {
		pciCompliance.LogAction("PROCESS_PAYMENT", "", txID, "FAILED", "Transaction not found")
		return
	}

	// Verify transaction signature
	txData := map[string]interface{}{
		"id":        txID,
		"amount":    tx.Amount.Value,
		"currency":  tx.Amount.Currency,
		"method":    tx.Method,
		"timestamp": tx.CreatedAt.Unix(),
	}

	if signature, ok := tx.Metadata["signature"].(string); ok {
		if !VerifyTransactionSignature(txData, signature) {
			h.UpdateTransactionStatus(txID, "failed")
			pciCompliance.LogAction("PROCESS_PAYMENT", "", txID, "FAILED", "Invalid signature")
			return
		}
	}

	// Update status to completed
	h.UpdateTransactionStatus(txID, "completed")

	// Create secure payment response
	h.mu.Lock()
	if tx, exists := h.transactions[txID]; exists {
		// Generate response token
		responseToken := generateSecureToken()

		tx.Response = &PaymentResponse{
			RequestID:  tx.Request.PaymentRequestID,
			MethodName: tx.Method,
			Details: map[string]interface{}{
				"transactionId": txID,
				"status":        "success",
				"token":         responseToken,
				"timestamp":     time.Now().Unix(),
			},
		}

		// Log successful payment
		pciCompliance.LogAction("PROCESS_PAYMENT", "", txID, "SUCCESS", tx.Request.Origin)
	}
	h.mu.Unlock()
}

// SerializePaymentRequest serializes a payment request from JSON
func SerializePaymentRequest(data []byte) (*PaymentRequest, error) {
	var req PaymentRequest
	err := json.Unmarshal(data, &req)
	return &req, err
}

// SerializePaymentResponse serializes a payment response to JSON
func SerializePaymentResponse(resp *PaymentResponse) ([]byte, error) {
	return json.Marshal(resp)
}
