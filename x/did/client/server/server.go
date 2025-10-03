// Package server provides a spawnable HTTP server for Auth service.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Errors
var (
	ErrAuthServerAlreadyRunning = errors.New("auth server already running")
	ErrAuthServerNotRunning     = errors.New("auth server not running")
	ErrFailedToStartAuthServer  = errors.New("failed to start auth server")
)

// AuthServer is a spawnable HTTP server for Auth service.
type AuthServer struct {
	*echo.Echo
	Port             int
	KillChan         chan bool
	ctx              context.Context
	cancel           context.CancelFunc
	sessionStore     map[string]string        // In-memory session store for WebAuthn challenges
	registrationDone chan error               // Channel to signal registration completion
	credentialData   chan *WebAuthnCredential // Channel to pass credential data to CLI
	username         string                   // Current username being registered
}

var authServer *AuthServer

// StartAuthServer starts the auth server
func StartAuthServer() error {
	if authServer != nil {
		return ErrAuthServerAlreadyRunning
	}
	setupAuthServer()
	return authServer.Start()
}

// StartAuthServerWithWebAuthn starts the auth server with WebAuthn support
func StartAuthServerWithWebAuthn(port int, username string, done chan error) error {
	if authServer != nil {
		return ErrAuthServerAlreadyRunning
	}
	setupAuthServerWithWebAuthn(port, username, done)
	return authServer.Start()
}

// StartAuthServerWithWebAuthnAndCredentialChannel starts auth server with WebAuthn and credential data channel
func StartAuthServerWithWebAuthnAndCredentialChannel(
	port int,
	username string,
	done chan error,
	credentialData chan *WebAuthnCredential,
) error {
	if authServer != nil {
		return ErrAuthServerAlreadyRunning
	}
	setupAuthServerWithWebAuthnAndCredentialChannel(port, username, done, credentialData)
	return authServer.Start()
}

// StartAuthServerForLogin starts the auth server for WebAuthn login
func StartAuthServerForLogin(port int, username string, done chan error) error {
	if authServer != nil {
		return ErrAuthServerAlreadyRunning
	}
	setupAuthServerForLogin(port, username, done)
	return authServer.Start()
}

// StopAuthServer stops the auth server
func StopAuthServer() error {
	if authServer == nil {
		return ErrAuthServerNotRunning
	}
	return authServer.Stop()
}

func (s *AuthServer) Start() error {
	// Setup signal context
	s.ctx, s.cancel = signal.NotifyContext(context.Background(), os.Interrupt)

	// Start server in goroutine
	go func() {
		if err := s.Echo.Start(fmt.Sprintf(":%d", s.Port)); err != nil &&
			err != http.ErrServerClosed {
			s.Logger.Fatal("shutting down the server")
		}
	}()

	// Start kill signal handler in another goroutine
	go s.HandleKillSignal()

	return nil
}

func (s *AuthServer) Stop() error {
	// Cancel the signal context to trigger shutdown
	if s.cancel != nil {
		s.cancel()
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		s.Logger.Fatal(err)
		return err
	}

	// Clean up
	destroyAuthServer()
	return nil
}

func (s *AuthServer) HandleKillSignal() {
	select {
	case <-s.KillChan:
		// Manual stop via KillChan
		s.Stop()
	case <-s.ctx.Done():
		// OS interrupt signal received
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			s.Logger.Fatal(err)
		}
		destroyAuthServer()
	}
}

// ╭───────────────────────────────────────────────────────────╮
// │                       Server Config                       │
// ╰───────────────────────────────────────────────────────────╯

func setupRoutes(e *echo.Echo) {
	// Basic routes
	e.GET("/", HandleIndex)
	e.GET("/health", HandleHealth)
	e.POST("/login", HandleLogin)

	// WebAuthn registration routes
	e.GET("/register", HandleWebAuthnRegister)
	e.GET("/begin-register", HandleBeginRegister)  // GET for fetching options
	e.POST("/begin-register", HandleBeginRegister) // POST also supported for client compatibility
	e.POST("/finish-register", HandleFinishRegister)
}

// setupMiddleware configures server middleware
func setupMiddleware(e *echo.Echo) {
	// CORS middleware for browser compatibility
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:*", "https://localhost:*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	// Security middleware
	e.Use(middleware.Secure())
	e.Use(middleware.RequestID())
	// Disable HTTP request logging for cleaner CLI output
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())
}

// destroyAuthServer destroys the auth server
func destroyAuthServer() {
	authServer = nil
}

// setupAuthServer sets up the auth server
func setupAuthServer() {
	authServer = &AuthServer{
		Echo:     echo.New(),
		Port:     8080,
		KillChan: make(chan bool),
	}
	// Disable Echo framework logging for cleaner CLI output
	authServer.HideBanner = true
	authServer.HidePort = true
	setupMiddleware(authServer.Echo)
	setupRoutes(authServer.Echo)
}

// setupAuthServerWithWebAuthn sets up the auth server with WebAuthn context
func setupAuthServerWithWebAuthn(port int, username string, done chan error) {
	// Initialize database for WebAuthn credential storage
	_ = InitDB() // Errors handled gracefully in storeWebAuthnCredential

	authServer = &AuthServer{
		Echo:             echo.New(),
		Port:             port,
		KillChan:         make(chan bool),
		sessionStore:     make(map[string]string),
		registrationDone: done,
		username:         username,
	}
	// Disable Echo framework logging for cleaner CLI output
	authServer.HideBanner = true
	authServer.HidePort = true
	setupMiddleware(authServer.Echo)
	setupRoutes(authServer.Echo)

	// Set up automatic server shutdown after 15 seconds as failsafe
	go func() {
		time.Sleep(15 * time.Second)
		if authServer != nil {
			logger := authServer.Logger
			logger.Warn("Auto-shutting down auth server after 15 second timeout")
			select {
			case authServer.KillChan <- true:
				logger.Info("Server shutdown signal sent via KillChan")
			default:
				logger.Warn("KillChan full, server may already be shutting down")
			}
		}
	}()
}

// setupAuthServerWithWebAuthnAndCredentialChannel sets up auth server with WebAuthn and credential channel
func setupAuthServerWithWebAuthnAndCredentialChannel(
	port int,
	username string,
	done chan error,
	credentialData chan *WebAuthnCredential,
) {
	// Initialize database for WebAuthn credential storage
	_ = InitDB() // Errors handled gracefully in storeWebAuthnCredential

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	authServer = &AuthServer{
		Echo:             e,
		Port:             port,
		KillChan:         make(chan bool),
		sessionStore:     make(map[string]string),
		registrationDone: done,
		credentialData:   credentialData,
		username:         username,
	}
	// Disable Echo framework logging for cleaner CLI output
	authServer.HideBanner = true
	authServer.HidePort = true
	setupMiddleware(authServer.Echo)
	setupRoutes(authServer.Echo)

	// Set up automatic server shutdown after 15 seconds as failsafe
	go func() {
		time.Sleep(15 * time.Second)
		if authServer != nil {
			logger := authServer.Logger
			logger.Warn("Auto-shutting down auth server after 15 second timeout")
			select {
			case authServer.KillChan <- true:
				logger.Info("Server shutdown signal sent via KillChan")
			default:
				logger.Warn("KillChan full, server may already be shutting down")
			}
		}
	}()
}

// setupAuthServerForLogin sets up the auth server for WebAuthn login
func setupAuthServerForLogin(port int, username string, done chan error) {
	// Initialize database for WebAuthn credential verification
	_ = InitDB() // Errors handled gracefully in login handlers

	authServer = &AuthServer{
		Echo:             echo.New(),
		Port:             port,
		KillChan:         make(chan bool),
		sessionStore:     make(map[string]string),
		registrationDone: done,
		username:         username,
	}
	// Disable Echo framework logging for cleaner CLI output
	authServer.HideBanner = true
	authServer.HidePort = true
	setupMiddleware(authServer.Echo)
	setupLoginRoutes(authServer.Echo)

	// Set up automatic server shutdown after 45 seconds as failsafe (longer for login)
	go func() {
		time.Sleep(45 * time.Second)
		if authServer != nil {
			logger := authServer.Logger
			logger.Warn("Auto-shutting down login auth server after 45 second timeout")
			select {
			case authServer.KillChan <- true:
				logger.Info("Login server shutdown signal sent via KillChan")
			default:
				logger.Warn("KillChan full, login server may already be shutting down")
			}
		}
	}()
}

// setupLoginRoutes configures routes specifically for login flow
func setupLoginRoutes(e *echo.Echo) {
	// Basic routes
	e.GET("/", HandleIndex)
	e.GET("/health", HandleHealth)

	// WebAuthn login routes
	e.GET("/login", HandleWebAuthnLogin)
	e.GET("/begin-login", HandleBeginLogin)
	e.POST("/begin-login", HandleBeginLogin) // POST also supported for client compatibility
	e.POST("/finish-login", HandleFinishLogin)
	e.POST("/login/verify", HandleFinishLogin) // Alternative endpoint for client compatibility
}
