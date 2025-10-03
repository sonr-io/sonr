package auth

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/keys"
	"github.com/sonr-io/sonr/client/tx"
	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// MockTxBuilder implements tx.TxBuilder for testing.
type MockTxBuilder struct {
	messages []sdk.Msg
	config   *tx.TxConfig
}

func NewMockTxBuilder() *MockTxBuilder {
	return &MockTxBuilder{
		messages: make([]sdk.Msg, 0),
		config: &tx.TxConfig{
			ChainID:       "test-chain",
			GasPrice:      0.001,
			GasDenom:      "usnr",
			GasLimit:      200000,
			GasAdjustment: 1.5,
		},
	}
}

func (m *MockTxBuilder) WithChainID(chainID string) tx.TxBuilder {
	m.config.ChainID = chainID
	return m
}

func (m *MockTxBuilder) WithGasPrice(price float64, denom string) tx.TxBuilder {
	m.config.GasPrice = price
	m.config.GasDenom = denom
	return m
}

func (m *MockTxBuilder) WithGasLimit(limit uint64) tx.TxBuilder {
	m.config.GasLimit = limit
	return m
}

func (m *MockTxBuilder) WithMemo(memo string) tx.TxBuilder {
	m.config.Memo = memo
	return m
}

func (m *MockTxBuilder) WithTimeoutHeight(height uint64) tx.TxBuilder {
	m.config.TimeoutHeight = height
	return m
}

func (m *MockTxBuilder) AddMessage(msg sdk.Msg) tx.TxBuilder {
	m.messages = append(m.messages, msg)
	return m
}

func (m *MockTxBuilder) AddMessages(msgs ...sdk.Msg) tx.TxBuilder {
	m.messages = append(m.messages, msgs...)
	return m
}

func (m *MockTxBuilder) ClearMessages() tx.TxBuilder {
	m.messages = make([]sdk.Msg, 0)
	return m
}

func (m *MockTxBuilder) WithFee(amount sdk.Coins) tx.TxBuilder {
	m.config.Fee = amount
	return m
}

func (m *MockTxBuilder) WithGasAdjustment(adjustment float64) tx.TxBuilder {
	m.config.GasAdjustment = adjustment
	return m
}

func (m *MockTxBuilder) EstimateGas(ctx context.Context) (uint64, error) {
	return 200000, nil
}

func (m *MockTxBuilder) Sign(ctx context.Context, keyring keys.KeyringManager) (*tx.SignedTx, error) {
	unsignedTx, _ := m.Build()
	return &tx.SignedTx{
		UnsignedTx: unsignedTx,
		Signature:  []byte("mock-signature"),
		PubKey:     []byte("mock-pubkey"),
		TxBytes:    []byte("mock-tx-bytes"),
	}, nil
}

func (m *MockTxBuilder) SignAndBroadcast(ctx context.Context, keyring keys.KeyringManager) (*tx.BroadcastResult, error) {
	return &tx.BroadcastResult{
		TxHash:    "mock-tx-hash",
		Height:    12345,
		Code:      0,
		Log:       "success",
		GasUsed:   100000,
		GasWanted: 200000,
	}, nil
}

func (m *MockTxBuilder) Broadcast(ctx context.Context, signedTx *tx.SignedTx) (*tx.BroadcastResult, error) {
	return &tx.BroadcastResult{
		TxHash:    "mock-tx-hash",
		Height:    12345,
		Code:      0,
		Log:       "success",
		GasUsed:   100000,
		GasWanted: 200000,
	}, nil
}

func (m *MockTxBuilder) Simulate(ctx context.Context) (*tx.SimulateResult, error) {
	return &tx.SimulateResult{
		GasWanted: 200000,
		GasUsed:   100000,
		Log:       "simulation success",
	}, nil
}

func (m *MockTxBuilder) Build() (*tx.UnsignedTx, error) {
	return &tx.UnsignedTx{
		Messages:  m.messages,
		Config:    m.config,
		SignBytes: []byte("mock-sign-bytes"),
	}, nil
}

func (m *MockTxBuilder) BuildSigned(signature []byte, pubKey []byte) (*tx.SignedTx, error) {
	unsignedTx, _ := m.Build()
	return &tx.SignedTx{
		UnsignedTx: unsignedTx,
		Signature:  signature,
		PubKey:     pubKey,
		TxBytes:    append(unsignedTx.SignBytes, signature...),
	}, nil
}

func (m *MockTxBuilder) Config() *tx.TxConfig {
	return m.config
}

// MockBroadcaster implements tx.Broadcaster for testing.
type MockBroadcaster struct {
	broadcastedTxs [][]byte
	shouldFail     bool
}

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		broadcastedTxs: make([][]byte, 0),
		shouldFail:     false,
	}
}

func (m *MockBroadcaster) Broadcast(ctx context.Context, txBytes []byte, mode tx.BroadcastMode) (*tx.BroadcastResult, error) {
	m.broadcastedTxs = append(m.broadcastedTxs, txBytes)

	if m.shouldFail {
		return &tx.BroadcastResult{
			Code: 1,
			Log:  "mock error",
		}, nil
	}

	return &tx.BroadcastResult{
		TxHash:    "mock-tx-hash",
		Height:    12345,
		Code:      0,
		Log:       "success",
		GasUsed:   100000,
		GasWanted: 200000,
	}, nil
}

func (m *MockBroadcaster) BroadcastSync(ctx context.Context, txBytes []byte) (*tx.BroadcastResult, error) {
	return m.Broadcast(ctx, txBytes, tx.BroadcastModeSync)
}

func (m *MockBroadcaster) BroadcastAsync(ctx context.Context, txBytes []byte) (*tx.BroadcastResult, error) {
	return m.Broadcast(ctx, txBytes, tx.BroadcastModeAsync)
}

func (m *MockBroadcaster) BroadcastBlock(ctx context.Context, txBytes []byte) (*tx.BroadcastResult, error) {
	return m.Broadcast(ctx, txBytes, tx.BroadcastModeBlock)
}

func (m *MockBroadcaster) BroadcastWithRetry(ctx context.Context, txBytes []byte, mode tx.BroadcastMode, maxRetries int) (*tx.BroadcastResult, error) {
	return m.Broadcast(ctx, txBytes, mode)
}

func (m *MockBroadcaster) WaitForConfirmation(ctx context.Context, txHash string, timeout time.Duration) (*tx.TxConfirmation, error) {
	return &tx.TxConfirmation{
		TxHash:      txHash,
		BlockHeight: 12345,
		BlockTime:   time.Now(),
		Code:        0,
		Log:         "confirmed",
		GasWanted:   200000,
		GasUsed:     100000,
	}, nil
}

func (m *MockBroadcaster) WithRetryConfig(config tx.RetryConfig) tx.Broadcaster {
	return m
}

func (m *MockBroadcaster) WithTimeout(timeout time.Duration) tx.Broadcaster {
	return m
}

// GaslessTestSuite tests gasless transaction functionality.
type GaslessTestSuite struct {
	suite.Suite
	manager     GaslessTransactionManager
	txBuilder   tx.TxBuilder
	broadcaster tx.Broadcaster
	config      *config.NetworkConfig
}

func (suite *GaslessTestSuite) SetupTest() {
	cfg := config.LocalNetwork()
	suite.config = &cfg
	suite.txBuilder = NewMockTxBuilder()
	suite.broadcaster = NewMockBroadcaster()
	suite.manager = NewGaslessTransactionManager(
		suite.txBuilder,
		suite.broadcaster,
		suite.config,
	)
}

func (suite *GaslessTestSuite) TestCreateGaslessRegistration() {
	// Create test WebAuthn credential
	credential := &WebAuthnCredential{
		ID:              "test-credential-id",
		RawID:           []byte("test-raw-id"),
		PublicKey:       []byte("test-public-key"),
		AttestationType: "none",
		Transports:      []string{"usb"},
		Flags: &AuthenticatorFlags{
			UserPresent:  true,
			UserVerified: true,
		},
		Authenticator: &AuthenticatorData{
			RPIDHash: []byte("test-rp-id-hash"),
		},
	}

	// Create options
	opts := &GaslessRegistrationOptions{
		Username:          "testuser",
		AutoCreateVault:   true,
		WebAuthnChallenge: []byte("test-challenge"),
	}

	// Create gasless registration
	gaslessTx, err := suite.manager.CreateGaslessRegistration(context.Background(), credential, opts)
	suite.Require().NoError(err)
	suite.Require().NotNil(gaslessTx)

	// Verify transaction fields
	suite.Equal("WebAuthn Gasless Registration", gaslessTx.Memo)
	suite.Equal(uint64(200000), gaslessTx.GasLimit)
	suite.Equal(signing.SignMode_SIGN_MODE_DIRECT, gaslessTx.SignMode)
	suite.Len(gaslessTx.Messages, 1)

	// Verify message type
	msg, ok := gaslessTx.Messages[0].(*didtypes.MsgRegisterWebAuthnCredential)
	suite.Require().True(ok)
	suite.Equal("testuser", msg.Username)
	suite.True(msg.AutoCreateVault)
}

func (suite *GaslessTestSuite) TestBroadcastGasless() {
	// Create test transaction
	gaslessTx := &GaslessTransaction{
		Messages: []sdk.Msg{
			&didtypes.MsgRegisterWebAuthnCredential{
				Controller: "sonr1xyz...",
				Username:   "testuser",
			},
		},
		Memo:          "Test Gasless",
		GasLimit:      200000,
		SignerAddress: "sonr1xyz...",
		SignMode:      signing.SignMode_SIGN_MODE_DIRECT,
		TxBytes:       []byte("test-tx-bytes"),
	}

	// Broadcast transaction
	result, err := suite.manager.BroadcastGasless(context.Background(), gaslessTx)
	suite.Require().NoError(err)
	suite.Require().NotNil(result)

	// Verify result
	suite.Equal("mock-tx-hash", result.TxHash)
	suite.Equal(int64(12345), result.Height)
	suite.Equal(uint32(0), result.Code)
}

func (suite *GaslessTestSuite) TestIsEligibleForGasless() {
	// Test WebAuthn registration message
	webauthnMsg := &didtypes.MsgRegisterWebAuthnCredential{
		Controller: "sonr1xyz...",
		Username:   "testuser",
	}

	// Should be eligible
	suite.True(suite.manager.IsEligibleForGasless([]sdk.Msg{webauthnMsg}))

	// Multiple messages should not be eligible
	suite.False(suite.manager.IsEligibleForGasless([]sdk.Msg{webauthnMsg, webauthnMsg}))

	// Other message types should not be eligible
	otherMsg := &didtypes.MsgCreateDID{
		Controller: "sonr1xyz...",
	}
	suite.False(suite.manager.IsEligibleForGasless([]sdk.Msg{otherMsg}))
}

func (suite *GaslessTestSuite) TestEstimateGaslessGas() {
	// Test WebAuthn registration gas estimate
	gas := suite.manager.EstimateGaslessGas("/did.v1.MsgRegisterWebAuthnCredential")
	suite.Equal(uint64(200000), gas)

	// Test default gas estimate
	gas = suite.manager.EstimateGaslessGas("/unknown.message.type")
	suite.Equal(uint64(100000), gas)
}

func TestGaslessTestSuite(t *testing.T) {
	suite.Run(t, new(GaslessTestSuite))
}

// TestValidateGaslessEligibility tests credential validation.
func TestValidateGaslessEligibility(t *testing.T) {
	tests := []struct {
		name       string
		credential *WebAuthnCredential
		wantError  bool
	}{
		{
			name: "valid credential",
			credential: &WebAuthnCredential{
				RawID:     []byte("test-id"),
				PublicKey: []byte("test-key"),
				Flags: &AuthenticatorFlags{
					UserPresent: true,
				},
			},
			wantError: false,
		},
		{
			name:       "nil credential",
			credential: nil,
			wantError:  true,
		},
		{
			name: "empty credential ID",
			credential: &WebAuthnCredential{
				PublicKey: []byte("test-key"),
				Flags: &AuthenticatorFlags{
					UserPresent: true,
				},
			},
			wantError: true,
		},
		{
			name: "empty public key",
			credential: &WebAuthnCredential{
				RawID: []byte("test-id"),
				Flags: &AuthenticatorFlags{
					UserPresent: true,
				},
			},
			wantError: true,
		},
		{
			name: "no user presence",
			credential: &WebAuthnCredential{
				RawID:     []byte("test-id"),
				PublicKey: []byte("test-key"),
				Flags: &AuthenticatorFlags{
					UserPresent: false,
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGaslessEligibility(tt.credential)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestGetGaslessEndpoint tests endpoint retrieval.
func TestGetGaslessEndpoint(t *testing.T) {
	cfg := &config.NetworkConfig{
		RPC: "http://localhost:26657",
	}

	endpoint := GetGaslessEndpoint(cfg)
	require.Equal(t, "http://localhost:26657", endpoint)
}

// TestWebAuthnCredentialConversion tests conversion between credential types.
func TestWebAuthnCredentialConversion(t *testing.T) {
	// Create client credential
	clientCred := &WebAuthnCredential{
		ID:              base64.RawURLEncoding.EncodeToString([]byte("test-id")),
		RawID:           []byte("test-id"),
		PublicKey:       []byte("test-public-key"),
		AttestationType: "none",
		Transports:      []string{"usb", "nfc"},
		Flags: &AuthenticatorFlags{
			UserPresent:  true,
			UserVerified: false,
		},
	}

	// Convert to protobuf credential
	protoCred := didtypes.WebAuthnCredential{
		CredentialId:    base64.RawURLEncoding.EncodeToString(clientCred.RawID),
		PublicKey:       clientCred.PublicKey,
		AttestationType: clientCred.AttestationType,
		Origin:          "http://localhost",
		Algorithm:       -7, // ES256
		CreatedAt:       time.Now().Unix(),
		RpId:            "localhost",
		RpName:          "Sonr Local",
		Transports:      clientCred.Transports,
		UserVerified:    clientCred.Flags.UserVerified,
	}

	// Verify conversion
	require.Equal(t, base64.RawURLEncoding.EncodeToString([]byte("test-id")), protoCred.CredentialId)
	require.Equal(t, clientCred.PublicKey, protoCred.PublicKey)
	require.Equal(t, clientCred.AttestationType, protoCred.AttestationType)
	require.Equal(t, clientCred.Transports, protoCred.Transports)
	require.Equal(t, clientCred.Flags.UserVerified, protoCred.UserVerified)
}
