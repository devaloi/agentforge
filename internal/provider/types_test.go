package provider

import (
	"encoding/json"
	"testing"
)

func TestMessageJSON(t *testing.T) {
	msg := Message{
		Role:    RoleUser,
		Content: "Hello, world",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Role != RoleUser {
		t.Errorf("Role = %q, want %q", decoded.Role, RoleUser)
	}
	if decoded.Content != "Hello, world" {
		t.Errorf("Content = %q, want %q", decoded.Content, "Hello, world")
	}
}

func TestToolCallParseArguments(t *testing.T) {
	tc := ToolCall{
		ID:        "call_123",
		Name:      "web_search",
		Arguments: `{"query":"Go REST API"}`,
	}

	args, err := tc.ParseArguments()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	q, ok := args["query"]
	if !ok {
		t.Fatal("missing key 'query'")
	}
	if q != "Go REST API" {
		t.Errorf("query = %q, want %q", q, "Go REST API")
	}
}

func TestToolCallParseArgumentsInvalid(t *testing.T) {
	tc := ToolCall{
		ID:        "call_123",
		Name:      "web_search",
		Arguments: `{invalid}`,
	}

	_, err := tc.ParseArguments()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestMessageWithToolCalls(t *testing.T) {
	msg := Message{
		Role:    RoleAssistant,
		Content: "",
		ToolCalls: []ToolCall{
			{ID: "call_1", Name: "web_search", Arguments: `{"query":"test"}`},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.ToolCalls) != 1 {
		t.Fatalf("ToolCalls len = %d, want 1", len(decoded.ToolCalls))
	}
	if decoded.ToolCalls[0].Name != "web_search" {
		t.Errorf("ToolCall name = %q, want %q", decoded.ToolCalls[0].Name, "web_search")
	}
}

func TestToolResultJSON(t *testing.T) {
	result := ToolResult{
		ToolCallID: "call_1",
		Content:    "search results here",
		IsError:    false,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ToolResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ToolCallID != "call_1" {
		t.Errorf("ToolCallID = %q, want %q", decoded.ToolCallID, "call_1")
	}
	if decoded.IsError {
		t.Error("IsError = true, want false")
	}
}

func TestUsageTotal(t *testing.T) {
	u := Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Usage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want 150", decoded.TotalTokens)
	}
}
