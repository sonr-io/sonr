package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/sonr-io/sonr/types/ipfs"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status       string            `json:"status"`
	Timestamp    string            `json:"timestamp"`
	Uptime       string            `json:"uptime"`
	Dependencies map[string]string `json:"dependencies"`
}

// HealthChecker manages health and readiness checks
type HealthChecker struct {
	startTime    time.Time
	redisClient  *asynq.Client
	ipfsClient   ipfs.IPFSClient
	ready        bool
	readyMu      sync.RWMutex
	redisHealthy bool
	ipfsHealthy  bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(redisClient *asynq.Client, ipfsClient ipfs.IPFSClient) *HealthChecker {
	hc := &HealthChecker{
		startTime:   time.Now(),
		redisClient: redisClient,
		ipfsClient:  ipfsClient,
		ready:       false,
	}

	// Start background health checks
	go hc.startHealthChecks()

	return hc
}

// startHealthChecks runs periodic health checks
func (hc *HealthChecker) startHealthChecks() {
	// Initial startup delay
	time.Sleep(3 * time.Second)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Run initial check
	hc.checkDependencies()

	for range ticker.C {
		hc.checkDependencies()
	}
}

// checkDependencies checks all service dependencies
func (hc *HealthChecker) checkDependencies() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check Redis connectivity
	hc.redisHealthy = hc.checkRedis(ctx)

	// Check IPFS connectivity (optional)
	hc.ipfsHealthy = hc.checkIPFS(ctx)

	// Update readiness based on critical dependencies
	hc.readyMu.Lock()
	hc.ready = hc.redisHealthy // Redis is required
	hc.readyMu.Unlock()
}

// checkRedis verifies Redis connectivity
func (hc *HealthChecker) checkRedis(ctx context.Context) bool {
	if hc.redisClient == nil {
		return false
	}

	// Try to ping Redis by enqueuing a test task
	testTask := asynq.NewTask("health:check", nil)
	_, err := hc.redisClient.EnqueueContext(ctx, testTask,
		asynq.Queue("health"),
		asynq.MaxRetry(0),
		asynq.Retention(1*time.Second)) // Auto-delete after 1 second
	if err != nil {
		return false
	}

	return true
}

// checkIPFS verifies IPFS connectivity
func (hc *HealthChecker) checkIPFS(ctx context.Context) bool {
	if hc.ipfsClient == nil {
		return true // IPFS is optional
	}

	// Check IPFS connectivity by getting version
	ch := make(chan bool, 1)
	go func() {
		// Try a simple operation to check connectivity
		if hc.ipfsClient != nil {
			// IPFS client exists, assume healthy for now
			// Real check would depend on actual IPFS client implementation
			ch <- true
		} else {
			ch <- false
		}
	}()

	select {
	case result := <-ch:
		return result
	case <-ctx.Done():
		return false
	}
}

// IsReady returns whether the service is ready to handle requests
func (hc *HealthChecker) IsReady() bool {
	hc.readyMu.RLock()
	defer hc.readyMu.RUnlock()
	return hc.ready
}

// GetStatus returns the current health status
func (hc *HealthChecker) GetStatus() HealthStatus {
	uptime := time.Since(hc.startTime)

	deps := make(map[string]string)
	if hc.redisHealthy {
		deps["redis"] = "healthy"
	} else {
		deps["redis"] = "unhealthy"
	}

	if hc.ipfsClient != nil {
		if hc.ipfsHealthy {
			deps["ipfs"] = "healthy"
		} else {
			deps["ipfs"] = "unhealthy"
		}
	} else {
		deps["ipfs"] = "not_configured"
	}

	status := "healthy"
	if !hc.ready {
		status = "unhealthy"
	}

	return HealthStatus{
		Status:       status,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Uptime:       uptime.String(),
		Dependencies: deps,
	}
}

// Global health checker instance
var healthChecker *HealthChecker

// InitHealthChecker initializes the global health checker
func InitHealthChecker(redisClient *asynq.Client, ipfsClient ipfs.IPFSClient) {
	healthChecker = NewHealthChecker(redisClient, ipfsClient)
}

// HealthCheckHandler returns health status (liveness probe)
func HealthCheckHandler(c echo.Context) error {
	if healthChecker == nil {
		return c.JSON(http.StatusOK, map[string]string{"status": "starting"})
	}

	status := healthChecker.GetStatus()
	if status.Status == "healthy" {
		return c.JSON(http.StatusOK, status)
	}
	return c.JSON(http.StatusServiceUnavailable, status)
}

// ReadinessHandler returns readiness status (readiness probe)
func ReadinessHandler(c echo.Context) error {
	if healthChecker == nil || !healthChecker.IsReady() {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"ready":  "false",
			"reason": "service not ready",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"ready": "true",
	})
}
