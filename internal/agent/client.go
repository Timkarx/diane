package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const defaultBaseURL = "http://localhost:4096"

func (p PromptResult) AsPlainText() []string {
	var text []string
	for _, part := range p.Parts {
		var meta struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(part.union, &meta); err != nil {
			slog.Error("decoding response part failed", "error", err)
			continue
		}

		switch meta.Type {
		case "text":
			textPart, err := part.AsTextPart()
			if err != nil {
				slog.Error("decode text part failed", "error", err)
				continue
			}
			text = append(text, textPart.Text)
			fmt.Println("text:", textPart.Text)
		case "tool":
			toolPart, err := part.AsToolPart()
			if err != nil {
				slog.Error("decode tool part failed", "error", err)
				continue
			}
			fmt.Printf("tool: %+v\n", toolPart)
		case "step-start":
			stepPart, err := part.AsStepStartPart()
			if err != nil {
				slog.Error("decode step start failed", "error", err)
				continue
			}
			fmt.Printf("step start: %+v\n", stepPart)
		}
	}
	return text
}

func (c *openCodeClient) CheckHealth() (HealthStatus, error) {
	slog.Info("req /global/health")

	var health HealthStatus
	if err := c.doJSON(http.MethodGet, "/global/health", nil, &health); err != nil {
		return HealthStatus{}, err
	}

	return health, nil
}

func (c *openCodeClient) Prompt(input string) (PromptResult, error) {

	res, err := c.prompt(input)
	if err != nil {
		slog.Error("prompt failed", "error", err)
		return PromptResult{}, err
	}
	return res, nil
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
