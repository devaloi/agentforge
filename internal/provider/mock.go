package provider

import (
	"context"
	"fmt"
	"sync"
)

// MockProvider is a deterministic provider for testing.
// It returns scripted responses in order of registration.
type MockProvider struct {
	mu        sync.Mutex
	responses []*Response
	index     int
	calls     []MockCall
}

// MockCall records a single call made to the mock provider.
type MockCall struct {
	Messages []Message
	Tools    []ToolDef
	Config   ChatConfig
}

// NewMockProvider creates a mock provider with the given scripted responses.
func NewMockProvider(responses ...*Response) *MockProvider {
	return &MockProvider{responses: responses}
}

// Name returns "mock".
func (m *MockProvider) Name() string { return "mock" }

// ChatComplete returns the next scripted response.
func (m *MockProvider) ChatComplete(_ context.Context, messages []Message, tools []ToolDef, cfg ChatConfig) (*Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, MockCall{
		Messages: messages,
		Tools:    tools,
		Config:   cfg,
	})

	if m.index >= len(m.responses) {
		return nil, fmt.Errorf("mock provider: no more scripted responses (called %d times, have %d responses)", m.index+1, len(m.responses))
	}

	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

// Calls returns all recorded calls.
func (m *MockProvider) Calls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.calls))
	copy(result, m.calls)
	return result
}

// Reset clears recorded calls and resets the response index.
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.index = 0
	m.calls = nil
}
