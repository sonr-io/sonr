package bridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHighwayService(t *testing.T) {
	service := NewHighwayService()

	require.NotNil(t, service)
	assert.NotNil(t, service.config)
	assert.NotNil(t, service.client)
	assert.NotNil(t, service.httpServer)
	assert.NotNil(t, service.queueManager)
	assert.NotNil(t, service.ctx)
	assert.NotNil(t, service.cancel)
	assert.NotNil(t, service.sigCh)
}

func TestHighwayServiceComponents(t *testing.T) {
	service := NewHighwayService()
	defer service.Shutdown()

	// Test that all components are properly initialized
	assert.NotNil(t, service.config.RedisAddr)
	assert.NotNil(t, service.config.JWTSecret)
	assert.NotNil(t, service.config.AsynqConfig)
}
