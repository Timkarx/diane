package opencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Timkarx/diane/core"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

type scheduleTaskMessage struct {
	Text string
}

func (m scheduleTaskMessage) ToText() string {
	return m.Text
}

type scheduleTaskOutput struct {
	ShouldNotify bool `json:"should_notify"`
}

type scheduleTaskSpec struct{}

func (scheduleTaskSpec) Schema() core.JSONSchema {
	return core.JSONSchema{
		"type": "object",
		"properties": map[string]any{
			"should_notify": map[string]any{
				"type": "boolean",
			},
		},
		"required": []string{"should_notify"},
	}
}

func TestNewOpenCodeClientDefaultsBaseURLAndPort(t *testing.T) {
	t.Helper()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{}, scheduleTaskSpec{})

	if client.baseURL != "http://localhost:4096" {
		t.Fatalf("baseURL = %q, want %q", client.baseURL, "http://localhost:4096")
	}
}

func TestNewOpenCodeClientAppliesConfiguredPortToBaseURL(t *testing.T) {
	t.Helper()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl: "http://example.com",
		Port:    5000,
	}, scheduleTaskSpec{})

	if client.baseURL != "http://example.com:5000" {
		t.Fatalf("baseURL = %q, want %q", client.baseURL, "http://example.com:5000")
	}
}

func TestNewOpenCodeClientPreservesBaseURLPortWhenPortUnset(t *testing.T) {
	t.Helper()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl: "http://example.com:1234",
	}, scheduleTaskSpec{})

	if client.baseURL != "http://example.com:1234" {
		t.Fatalf("baseURL = %q, want %q", client.baseURL, "http://example.com:1234")
	}
}

func TestNewOpenCodeClientOverridesExistingPortWhenConfigured(t *testing.T) {
	t.Helper()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl: "http://example.com:1234",
		Port:    5000,
	}, scheduleTaskSpec{})

	if client.baseURL != "http://example.com:5000" {
		t.Fatalf("baseURL = %q, want %q", client.baseURL, "http://example.com:5000")
	}
}

func (scheduleTaskSpec) ShouldAct() bool {
	return false
}

func (scheduleTaskSpec) ExecuteEffect(scheduleTaskMessage, scheduleTaskOutput) error {
	return nil
}

func (scheduleTaskSpec) Validate() error {
	return nil
}

func TestScheduleTaskReturnsStructuredOutput(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			if r.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}
			writeTestJSON(w, map[string]any{"id": "session-1"})
		case "/session/session-1/message":
			if r.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}
			writeTestJSON(w, map[string]any{
				"info": map[string]any{
					"structured": map[string]any{
						"should_notify": true,
					},
				},
				"parts": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:    server.URL,
		HTTPClient: server.Client(),
	}, scheduleTaskSpec{})

	got, err := client.ScheduleTask(scheduleTaskMessage{Text: "hello"})
	if err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}

	if !got.ShouldNotify {
		t.Fatalf("ScheduleTask() = %#v, want should_notify=true", got)
	}
}

func TestScheduleTaskDefaultModeCreatesNewSessionPerMessage(t *testing.T) {
	t.Helper()

	var sessionCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			if r.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}

			id := sessionCalls.Add(1)
			writeTestJSON(w, map[string]any{"id": fmt.Sprintf("session-%d", id)})
		case "/session/session-1/message", "/session/session-2/message":
			if r.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}

			writeTestJSON(w, map[string]any{
				"info":  map[string]any{"structured": map[string]any{"should_notify": true}},
				"parts": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:    server.URL,
		HTTPClient: server.Client(),
	}, scheduleTaskSpec{})

	for range 2 {
		if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "hello"}); err != nil {
			t.Fatalf("ScheduleTask() error = %v", err)
		}
	}

	if got := sessionCalls.Load(); got != 2 {
		t.Fatalf("session create calls = %d, want 2", got)
	}
}

func TestScheduleTaskReuseModeReusesSingleSession(t *testing.T) {
	t.Helper()

	var sessionCalls atomic.Int32
	var messageCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			if r.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}

			sessionCalls.Add(1)
			writeTestJSON(w, map[string]any{"id": "session-1"})
		case "/session/session-1/message":
			if r.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}

			messageCalls.Add(1)
			writeTestJSON(w, map[string]any{
				"info":  map[string]any{"structured": map[string]any{"should_notify": true}},
				"parts": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:     server.URL,
		HTTPClient:  server.Client(),
		SessionMode: core.TaskAgentSessionModeReusePerClient,
	}, scheduleTaskSpec{})

	for range 2 {
		if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "hello"}); err != nil {
			t.Fatalf("ScheduleTask() error = %v", err)
		}
	}

	if got := sessionCalls.Load(); got != 1 {
		t.Fatalf("session create calls = %d, want 1", got)
	}

	if got := messageCalls.Load(); got != 2 {
		t.Fatalf("message calls = %d, want 2", got)
	}
}

func TestScheduleTaskReuseModeRecreatesMissingSession(t *testing.T) {
	t.Helper()

	var sessionCalls atomic.Int32
	var messageCalls atomic.Int32
	var firstStaleAttempt atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			if r.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}

			id := sessionCalls.Add(1)
			if id == 1 {
				writeTestJSON(w, map[string]any{"id": "session-stale"})
				return
			}

			writeTestJSON(w, map[string]any{"id": "session-fresh"})
		case "/session/session-stale/message":
			messageCalls.Add(1)
			if firstStaleAttempt.CompareAndSwap(false, true) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, "unexpected retry to stale session", http.StatusInternalServerError)
		case "/session/session-fresh/message":
			messageCalls.Add(1)
			writeTestJSON(w, map[string]any{
				"info":  map[string]any{"structured": map[string]any{"should_notify": true}},
				"parts": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:     server.URL,
		HTTPClient:  server.Client(),
		SessionMode: core.TaskAgentSessionModeReusePerClient,
	}, scheduleTaskSpec{})

	if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "hello"}); err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}

	if got := sessionCalls.Load(); got != 2 {
		t.Fatalf("session create calls = %d, want 2", got)
	}

	if got := messageCalls.Load(); got != 2 {
		t.Fatalf("message calls = %d, want 2", got)
	}
}

func TestScheduleTaskReuseModeCreatesOneSessionForConcurrentCalls(t *testing.T) {
	t.Helper()

	var sessionCalls atomic.Int32
	var messageCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			sessionCalls.Add(1)
			writeTestJSON(w, map[string]any{"id": "session-1"})
		case "/session/session-1/message":
			messageCalls.Add(1)
			writeTestJSON(w, map[string]any{
				"info":  map[string]any{"structured": map[string]any{"should_notify": true}},
				"parts": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:     server.URL,
		HTTPClient:  server.Client(),
		SessionMode: core.TaskAgentSessionModeReusePerClient,
	}, scheduleTaskSpec{})

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "hello"}); err != nil {
				t.Errorf("ScheduleTask() error = %v", err)
			}
		}()
	}
	wg.Wait()

	if got := sessionCalls.Load(); got != 1 {
		t.Fatalf("session create calls = %d, want 1", got)
	}

	if got := messageCalls.Load(); got != 5 {
		t.Fatalf("message calls = %d, want 5", got)
	}
}

func TestScheduleTaskOmitsAgentAndModelByDefault(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			writeTestJSON(w, map[string]any{"id": "session-1"})
		case "/session/session-1/message":
			body := readPromptBody(t, r)
			if _, ok := body["agent"]; ok {
				t.Fatalf("request unexpectedly included agent: %#v", body["agent"])
			}
			if _, ok := body["model"]; ok {
				t.Fatalf("request unexpectedly included model: %#v", body["model"])
			}

			writeTestJSON(w, map[string]any{
				"info":  map[string]any{"structured": map[string]any{"should_notify": true}},
				"parts": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:    server.URL,
		HTTPClient: server.Client(),
	}, scheduleTaskSpec{})

	if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "hello"}); err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}
}

func TestScheduleTaskIncludesConfiguredAgentAndModel(t *testing.T) {
	t.Helper()

	type promptConfig struct {
		Agent string
		Model map[string]any
	}

	var seen []promptConfig
	var seenMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			writeTestJSON(w, map[string]any{"id": "session-1"})
		case "/session/session-1/message":
			body := readPromptBody(t, r)
			config := promptConfig{}
			if agent, ok := body["agent"].(string); ok {
				config.Agent = agent
			}
			if model, ok := body["model"].(map[string]any); ok {
				config.Model = model
			}

			seenMu.Lock()
			seen = append(seen, config)
			seenMu.Unlock()

			writeTestJSON(w, map[string]any{
				"info":  map[string]any{"structured": map[string]any{"should_notify": true}},
				"parts": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOpenCodeClient[scheduleTaskMessage, scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:    server.URL,
		HTTPClient: server.Client(),
	}, scheduleTaskSpec{})

	client.SetAgent("planner")
	client.SetModel("openai", "gpt-5")
	if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "hello"}); err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}

	client.SetAgent("coder")
	client.SetModel("anthropic", "claude-sonnet")
	if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "again"}); err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}

	client.ClearAgent()
	client.ClearModel()
	if _, err := client.ScheduleTask(scheduleTaskMessage{Text: "default"}); err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}

	seenMu.Lock()
	defer seenMu.Unlock()

	if len(seen) != 3 {
		t.Fatalf("captured %d requests, want 3", len(seen))
	}

	if seen[0].Agent != "planner" || seen[0].Model["providerID"] != "openai" || seen[0].Model["modelID"] != "gpt-5" {
		t.Fatalf("first request = %#v, want planner/openai/gpt-5", seen[0])
	}

	if seen[1].Agent != "coder" || seen[1].Model["providerID"] != "anthropic" || seen[1].Model["modelID"] != "claude-sonnet" {
		t.Fatalf("second request = %#v, want coder/anthropic/claude-sonnet", seen[1])
	}

	if seen[2].Agent != "" || seen[2].Model != nil {
		t.Fatalf("third request = %#v, want cleared agent/model", seen[2])
	}
}

func readPromptBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	return payload
}

func writeTestJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
