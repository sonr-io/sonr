package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/sonr-io/sonr/crypto/mpc"
)

// MockAsynqClient provides a test double for asynq.Client
type MockAsynqClient struct {
	enqueuedTasks []MockTask
}

type MockTask struct {
	Type    string
	Payload []byte
	Queue   string
}

func (m *MockAsynqClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	mockTask := MockTask{
		Type:    task.Type(),
		Payload: task.Payload(),
		Queue:   "default",
	}
	m.enqueuedTasks = append(m.enqueuedTasks, mockTask)

	return &asynq.TaskInfo{
		ID:    "test-task-id",
		Type:  task.Type(),
		Queue: mockTask.Queue,
	}, nil
}

func (m *MockAsynqClient) Close() error {
	return nil
}

// setupTestEcho creates a test Echo server and mock client for benchmarking
func setupTestEcho() (*echo.Echo, *MockAsynqClient) {
	mockClient := &MockAsynqClient{}
	config := &Config{
		JWTSecret:  []byte("test-secret"),
		IPFSClient: &MockIPFSClient{},
	}
	s := NewServer(config)
	return s.Echo(), mockClient
}

// BenchmarkHealthCheckHandler measures the performance of the health check endpoint
func BenchmarkHealthCheckHandler(b *testing.B) {
	e, _ := setupTestEcho()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/health", nil)
			rr := httptest.NewRecorder()
			e.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", rr.Code)
			}
		}
	})
}

// BenchmarkGenerateHandler measures the performance of the generate endpoint
func BenchmarkGenerateHandler(b *testing.B) {
	e, client := setupTestEcho()

	payload := map[string]any{
		"user_id":  123,
		"priority": "default",
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(payload)
	requestBody := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/vault/generate", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			e.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", rr.Code)
			}
		}
	})

	b.StopTimer()
	b.Logf("Tasks enqueued: %d", len(client.enqueuedTasks))
}

// BenchmarkSignHandler measures the performance of the sign endpoint
func BenchmarkSignHandler(b *testing.B) {
	e, client := setupTestEcho()

	payload := map[string]any{
		"message":  []byte("benchmark test message"),
		"enclave":  &mpc.EnclaveData{},
		"priority": "default",
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(payload)
	requestBody := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/vault/sign", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			e.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", rr.Code)
			}
		}
	})

	b.StopTimer()
	b.Logf("Tasks enqueued: %d", len(client.enqueuedTasks))
}

// BenchmarkMemoryAllocation measures memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	e, _ := setupTestEcho()

	payload := map[string]any{
		"user_id": 123,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(payload)
	requestBody := buf.Bytes()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ReportAllocs()

	for b.Loop() {
		req := httptest.NewRequest("POST", "/vault/generate", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, req)
	}

	b.StopTimer()
	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.Logf("Memory allocated per operation: %d bytes", (m2.TotalAlloc-m1.TotalAlloc)/uint64(b.N))
	b.Logf("Total allocations: %d", m2.Mallocs-m1.Mallocs)
}

// BenchmarkLatencyMeasurement measures end-to-end latency
func BenchmarkLatencyMeasurement(b *testing.B) {
	e, _ := setupTestEcho()

	payload := map[string]any{
		"user_id": 123,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(payload)
	requestBody := buf.Bytes()

	var totalLatency time.Duration
	minLatency := time.Hour
	var maxLatency time.Duration

	for b.Loop() {
		start := time.Now()

		req := httptest.NewRequest("POST", "/vault/generate", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, req)

		latency := time.Since(start)
		totalLatency += latency

		if latency < minLatency {
			minLatency = latency
		}
		if latency > maxLatency {
			maxLatency = latency
		}

		if rr.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", rr.Code)
		}
	}

	b.StopTimer()

	avgLatency := totalLatency / time.Duration(b.N)
	b.Logf("Average latency: %v", avgLatency)
	b.Logf("Min latency: %v", minLatency)
	b.Logf("Max latency: %v", maxLatency)
}
