package handlers

import (
	"encoding/json"
	"testing"

	"github.com/sonr-io/sonr/crypto/mpc"
)

// getQueueFromPriority returns the queue name based on priority
func getQueueFromPriority(priority string) string {
	return GetQueueFromPriority(priority)
}

// BenchmarkJSONMarshaling measures JSON encoding/decoding performance
func BenchmarkJSONMarshaling(b *testing.B) {
	payload := map[string]any{
		"message":  []byte("benchmark test message for JSON marshaling performance"),
		"enclave":  &mpc.EnclaveData{},
		"priority": "critical",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Marshal
			data, err := json.Marshal(payload)
			if err != nil {
				b.Error(err)
			}

			// Unmarshal
			var decoded map[string]any
			err = json.Unmarshal(data, &decoded)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkQueuePrioritySelection measures queue selection performance
func BenchmarkQueuePrioritySelection(b *testing.B) {
	priorities := []string{"critical", "high", "default", "low", "", "unknown"}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			priority := priorities[i%len(priorities)]
			i++
			queue := getQueueFromPriority(priority)
			_ = queue // Avoid compiler optimization
		}
	})
}
