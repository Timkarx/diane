package opencode

import (
	"diane/core"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func (scheduleTaskSpec) ShouldAct() bool {
	return false
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

	client := NewOpenCodeClient[scheduleTaskOutput, scheduleTaskSpec](core.TaskAgentOptions{
		BaseUrl:    server.URL,
		HTTPClient: server.Client(),
	})

	got, err := client.ScheduleTask(core.TaskAgentMessage{Text: "hello"})
	if err != nil {
		t.Fatalf("ScheduleTask() error = %v", err)
	}

	if !got.ShouldNotify {
		t.Fatalf("ScheduleTask() = %#v, want should_notify=true", got)
	}
}

func writeTestJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
