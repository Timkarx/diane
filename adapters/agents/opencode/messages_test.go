package opencode

import (
	"github.com/Timkarx/diane/core"
	"io"
	"os"
	"strings"
	"testing"
)

func TestOpencodeResultAsPlainTextIgnoresNonTextParts(t *testing.T) {
	result := OpencodeResult[core.Unstructured]{
		Parts: []Part{
			mustPart(t, `{"id":"text-1","messageID":"msg-1","sessionID":"session-1","text":"hello","type":"text"}`),
			mustPart(t, `{"id":"tool-1","callID":"call-1","messageID":"msg-1","sessionID":"session-1","state":{"status":"pending"},"tool":"bash","type":"tool"}`),
			mustPart(t, `{"id":"step-1","messageID":"msg-1","sessionID":"session-1","type":"step-start"}`),
			mustPart(t, `{"id":"text-2","messageID":"msg-1","sessionID":"session-1","text":"world","type":"text"}`),
		},
	}

	stdout := captureStdout(t, func() {
		got := result.AsPlainText()
		if strings.Join(got, " ") != "hello world" {
			t.Fatalf("AsPlainText() = %q, want %q", got, []string{"hello", "world"})
		}
	})

	if stdout != "" {
		t.Fatalf("AsPlainText() wrote to stdout: %q", stdout)
	}
}

func TestOpencodeResultDebugPrintPrettyPrintsAllParts(t *testing.T) {
	result := OpencodeResult[core.Unstructured]{
		Info: AssistantMessage{
			Agent:      "planner",
			Id:         "msg-1",
			Mode:       "chat",
			ModelID:    "model-1",
			ParentID:   "parent-1",
			ProviderID: "provider-1",
			Role:       "assistant",
			SessionID:  "session-1",
			Path: struct {
				Cwd  string `json:"cwd"`
				Root string `json:"root"`
			}{
				Cwd:  "/tmp",
				Root: "/",
			},
			Time: struct {
				Completed *float32 `json:"completed,omitempty"`
				Created   float32  `json:"created"`
			}{
				Created: 1,
			},
		},
		Parts: []Part{
			mustPart(t, `{"id":"text-1","messageID":"msg-1","sessionID":"session-1","text":"hello","type":"text"}`),
			mustPart(t, `{"id":"step-1","messageID":"msg-1","sessionID":"session-1","type":"step-start"}`),
		},
	}

	stdout := captureStdout(t, func() {
		result.DebugPrint()
	})

	for _, want := range []string{
		"info:",
		"part 0:",
		"part 1:",
		"\n  \"agent\": \"planner\"",
		"\n  \"type\": \"text\"",
		"\n  \"type\": \"step-start\"",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("DebugPrint() output missing %q in %q", want, stdout)
		}
	}
}

func mustPart(t *testing.T, raw string) Part {
	t.Helper()

	var part Part
	if err := part.UnmarshalJSON([]byte(raw)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	return part
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("stdout close error = %v", err)
	}

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if err := r.Close(); err != nil {
		t.Fatalf("stdout read close error = %v", err)
	}

	return string(output)
}
