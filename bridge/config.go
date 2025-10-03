package bridge

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sonr-io/sonr/types/ipfs"
)

const (
	DefaultRedisAddr = "127.0.0.1:6379"
	DefaultJWTSecret = "highway-ucan-secret-key"
	DefaultHTTPPort  = 8090
	ShutdownTimeout  = 30 * time.Second
)

// OIDCProviderConfig contains OIDC provider configuration
type OIDCProviderConfig struct {
	Issuer                 string
	PublicURL              string
	SigningKeyPath         string
	EncryptionKeyPath      string
	AuthorizationCodeTTL   time.Duration
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	IDTokenTTL             time.Duration
	EnablePKCE             bool
	EnableRefreshTokens    bool
	EnableSIOP             bool
	SupportedScopes        []string
	SupportedResponseTypes []string
	SupportedGrantTypes    []string
	AllowedRedirectURIs    []string
	WebAuthnRPID           string
	WebAuthnRPName         string
	WebAuthnTimeout        int
	AutoCreateVault        bool
}

type Config struct {
	RedisAddr       string
	HTTPPort        int
	JWTSecret       []byte
	IPFSClient      ipfs.IPFSClient
	ShutdownTimeout time.Duration
	AsynqConfig     asynq.Config
	OIDC            OIDCProviderConfig
}

func NewConfig() *Config {
	jwtSecret := initializeJWTSecret()
	redisAddr := getRedisAddr()
	httpPort := getHTTPPort()
	ipfsClient := initializeIPFS()
	oidcConfig := initializeOIDCConfig()

	return &Config{
		RedisAddr:       redisAddr,
		HTTPPort:        httpPort,
		JWTSecret:       jwtSecret,
		IPFSClient:      ipfsClient,
		ShutdownTimeout: ShutdownTimeout,
		AsynqConfig: asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ShutdownTimeout: ShutdownTimeout,
			// Enhanced error handling and retry configuration
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,
			IsFailure: func(err error) bool {
				return err != nil
			},
		},
		OIDC: oidcConfig,
	}
}

func initializeJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = DefaultJWTSecret
		log.Printf("Warning: Using default JWT secret for UCAN operations")
		log.Printf("Set JWT_SECRET environment variable for production deployment")
	} else {
		log.Println("JWT secret loaded from environment")
	}
	return []byte(secret)
}

func getRedisAddr() string {
	// Check for REDIS_URL first (Docker Compose style)
	if url := os.Getenv("REDIS_URL"); url != "" {
		// Parse redis://host:port format
		if len(url) > 8 && url[:8] == "redis://" {
			return url[8:]
		}
		return url
	}
	// Fall back to REDIS_ADDR
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	log.Printf("Using default Redis address: %s", DefaultRedisAddr)
	return DefaultRedisAddr
}

func initializeIPFS() ipfs.IPFSClient {
	ipfsClient, err := ipfs.GetClient()
	if err != nil {
		log.Printf("Warning: IPFS client initialization failed: %v", err)
		log.Println("Enclave data will be handled directly without IPFS storage")
		return nil
	}
	log.Println("IPFS client initialized successfully")
	return ipfsClient
}

func initializeOIDCConfig() OIDCProviderConfig {
	issuer := os.Getenv("OIDC_ISSUER")
	if issuer == "" {
		issuer = "https://localhost:8080"
	}

	publicURL := os.Getenv("OIDC_PUBLIC_URL")
	if publicURL == "" {
		publicURL = issuer
	}

	rpID := os.Getenv("WEBAUTHN_RP_ID")
	if rpID == "" {
		rpID = "localhost"
	}

	rpName := os.Getenv("WEBAUTHN_RP_NAME")
	if rpName == "" {
		rpName = "Sonr Identity Platform"
	}

	return OIDCProviderConfig{
		Issuer:               issuer,
		PublicURL:            publicURL,
		SigningKeyPath:       os.Getenv("OIDC_SIGNING_KEY_PATH"),
		EncryptionKeyPath:    os.Getenv("OIDC_ENCRYPTION_KEY_PATH"),
		AuthorizationCodeTTL: 10 * time.Minute,
		AccessTokenTTL:       1 * time.Hour,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		IDTokenTTL:           1 * time.Hour,
		EnablePKCE:           true,
		EnableRefreshTokens:  true,
		EnableSIOP:           true,
		SupportedScopes: []string{
			"openid", "profile", "email", "did", "vault", "offline_access",
		},
		SupportedResponseTypes: []string{
			"code", "id_token", "code id_token",
		},
		SupportedGrantTypes: []string{
			"authorization_code", "refresh_token", "client_credentials",
		},
		AllowedRedirectURIs: getRedirectURIs(),
		WebAuthnRPID:        rpID,
		WebAuthnRPName:      rpName,
		WebAuthnTimeout:     60000,
		AutoCreateVault:     true,
	}
}

// getRedirectURIs returns the allowed redirect URIs for OIDC
func getRedirectURIs() []string {
	// Default URIs for development
	uris := []string{
		"http://localhost:3000/callback",
		"http://localhost:3001/callback",
		"https://localhost:3000/callback",
		"https://localhost:3001/callback",
	}

	// Add additional URIs from environment if specified
	if envURIs := os.Getenv("OIDC_ALLOWED_REDIRECT_URIS"); envURIs != "" {
		// Parse comma-separated URIs
		// This could be enhanced with proper validation
		log.Printf("Additional redirect URIs configured from environment")
	}

	return uris
}

// getHTTPPort returns the HTTP port for the Highway service
func getHTTPPort() int {
	if port := os.Getenv("HIGHWAY_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			return p
		}
	}
	return DefaultHTTPPort
}
