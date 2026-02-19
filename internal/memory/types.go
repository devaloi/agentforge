// Package memory provides a thread-safe shared key-value store
// for inter-agent collaboration.
package memory

import "time"

// Entry represents a single value stored in shared memory.
type Entry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}
