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
	spec           *T
}

func (c *OpencodeAgent[K, T]) CheckHealth() (core.HealthStatus, error) {
	slog.Info("req /global/health")

	var health core.HealthStatus
	if err := c.doJSON(http.MethodGet, "/global/health", nil, &health); err != nil {
		return core.HealthStatus{}, err
	}

	return health, nil
}

func (c *OpencodeAgent[K, T]) ScheduleTask(message core.TaskAgentMessage) (K, error) {
	res, err := c.prompt(message)
	if err != nil {
		slog.Error("schedule task failed", "error", err)
		var zero K
		return zero, err
	}

	structured, err := res.Structured()
	if err != nil {
		slog.Error("decode structured output failed", "error", err)
		var zero K
		return zero, err
	}

	if c.spec != nil {
		if err := (*c.spec).ExecuteEffect(structured); err != nil {
			return structured, err
		}
	}

	return structured, nil
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
