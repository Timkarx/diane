package opencode

import (
	"diane/core"
	"log/slog"
	"net/http"
)

const defaultBaseURL = "http://localhost:4096"

type OpencodeAgent[K any, T core.TaskSpec[K]] struct {
	httpClient     *http.Client
	baseURL        string
	requestCounter int
}

func (c *OpencodeAgent[K, T]) CheckHealth() (core.HealthStatus, error) {
	slog.Info("req /global/health")

	var health core.HealthStatus
	if err := c.doJSON(http.MethodGet, "/global/health", nil, &health); err != nil {
		return core.HealthStatus{}, err
	}

	return health, nil
}

func (c *OpencodeAgent[K, T]) Prompt(message core.TaskAgentMessage) (OpencodeResult[T], error) {
	res, err := c.prompt(message)
	if err != nil {
		slog.Error("prompt failed", "error", err)
		return OpencodeResult[T]{}, err
	}
	return res, nil
}

func NewOpenCodeClient[K any, T core.TaskSpec[K]](opts core.TaskAgentOptions) *OpencodeAgent[K, T] {
	slog.Info("initializing opencode client")

	baseURL := opts.BaseUrl
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &OpencodeAgent[K, T]{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}
