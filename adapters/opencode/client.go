package opencode

import (
	"log/slog"
	"net/http"
	"diane/core"
)

const defaultBaseURL = "http://localhost:4096"

type OpencodeAgent[T core.Actionable] struct {
	httpClient     *http.Client
	baseURL        string
	requestCounter int
}

func (c *OpencodeAgent[T]) CheckHealth() (core.HealthStatus, error) {
	slog.Info("req /global/health")

	var health core.HealthStatus
	if err := c.doJSON(http.MethodGet, "/global/health", nil, &health); err != nil {
		return HealthStatus{}, err
	}

	return health, nil
}

func (c *openCodeClient[T]) Prompt(message ClientMessage) (PromptResult[T], error) {
	res, err := c.prompt(message)
	if err != nil {
		slog.Error("prompt failed", "error", err)
		return PromptResult[T]{}, err
	}
	return res, nil
}

func NewOpenCodeClient[T Actionable](opts ClientOptions) *openCodeClient[T] {
	slog.Info("initializing opencode client")

	baseURL := opts.BaseUrl
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &openCodeClient[T]{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}
