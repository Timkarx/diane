package main

import (
	"bytes"
	"diane/internal/agent"
	serverpkg "diane/internal/server"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	server := newIntegrationServer(t)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("get /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want %q", got, "application/json")
	}

	var body struct {
		Healthy bool `json:"healthy"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if !body.Healthy {
		t.Fatalf("healthy = %v, want true", body.Healthy)
	}
}

func TestPromptEndpointWithLiveBackend(t *testing.T) {
	requireLiveBackend(t)

	server := newIntegrationServer(t)
	defer server.Close()

	resp, err := server.Client().Post(
		server.URL+"/prompt",
		"application/json",
		bytes.NewBufferString(`{"prompt":"Reply with the exact word PONG and no other text."}`),
	)
	if err != nil {
		t.Fatalf("post /prompt: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d, body = %s", resp.StatusCode, http.StatusOK, strings.TrimSpace(string(body)))
	}

	var body struct {
		Text []string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode prompt response: %v", err)
	}

	joined := strings.TrimSpace(strings.Join(body.Text, "\n"))
	if joined == "" {
		t.Fatal("prompt text should not be empty")
	}
}

func TestPromptEndpointRejectsInvalidPayload(t *testing.T) {
	server := newIntegrationServer(t)
	defer server.Close()

	resp, err := server.Client().Post(
		server.URL+"/prompt",
		"application/json",
		bytes.NewBufferString(`{"prompt":123}`),
	)
	if err != nil {
		t.Fatalf("post /prompt with invalid payload: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode invalid payload response: %v", err)
	}

	if body.Error != "Invalid payload shape" {
		t.Fatalf("error = %q, want %q", body.Error, "Invalid payload shape")
	}
}

func newIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewUnstartedServer(serverpkg.New())
	server.Config.ReadHeaderTimeout = 5 * time.Second
	server.Start()

	t.Cleanup(server.Close)
	return server
}

func requireLiveBackend(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping live backend integration test in short mode")
	}

	client := agent.NewOpenCodeClient(agent.ClientOptions{})
	health, err := client.CheckHealth()
	if err != nil {
		t.Skipf("live backend unavailable on http://localhost:4096: %v", err)
	}

	if !health.Healthy {
		t.Skip("live backend reported unhealthy")
	}
}
