package agent

import (
	"log/slog"
	"net/http"
	"encoding/json"
	"fmt"
)

const defaultBaseURL = "http://localhost:4096"

func (c *openCodeClient) CheckHealth() (HealthStatus, error) {
	slog.Info("req /global/health")

	var health HealthStatus
	if err := c.doJSON(http.MethodGet, "/global/health", nil, &health); err != nil {
		return HealthStatus{}, err
	}

	return health, nil
}

func (c *openCodeClient) EvaluateInput(input string) {

	res, err := c.Prompt(input) 
	if err != nil {
		slog.Error("prompt failed", "error", err)
		return
	}

	b, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		slog.Error("prompt failed", "error", err)
		return
	}
	fmt.Println(string(b))

	return
}

func (c *openCodeClient) Prompt(input string) (PromptResult, error) {
	session, err := c.createSession()
	if err != nil {
		return PromptResult{}, err
	}

	return c.sendMessage(session.Id, input)
}

func NewOpenCodeClient(opts ClientOptions) *openCodeClient {
	slog.Info("initializing opencode client")

	baseURL := opts.BaseUrl
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &openCodeClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}
