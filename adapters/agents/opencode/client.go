package opencode

import (
	"github.com/Timkarx/diane/core"
	"log/slog"
	"net/http"
	"sync"
)

const defaultBaseURL = "http://localhost:4096"

type OpencodeAgent[S core.TaskAgentMessage, K any, T core.TaskSpec[S, K]] struct {
	httpClient     *http.Client
	baseURL        string
	requestCounter uint64
	sessionMode    core.TaskAgentSessionMode
	sessionMu      sync.Mutex
	sessionID      string
	optionsMu      sync.RWMutex
	agent          *string
	model          *modelConfig
	spec           *T
}

type modelConfig struct {
	providerID string
	modelID    string
}

func (c *OpencodeAgent[S, K, T]) CheckHealth() (core.HealthStatus, error) {
	slog.Info("req /global/health")

	var health core.HealthStatus
	if err := c.doJSON(http.MethodGet, "/global/health", nil, &health); err != nil {
		return core.HealthStatus{}, err
	}

	return health, nil
}

func (c *OpencodeAgent[S, K, T]) ScheduleTask(message S) (K, error) {
	res, err := c.prompt(message.ToText())
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
		if err := (*c.spec).ExecuteEffect(message, structured); err != nil {
			return structured, err
		}
	}

	return structured, nil
}

func NewOpenCodeClient[S core.TaskAgentMessage, K any, T core.TaskSpec[S, K]](opts core.TaskAgentOptions, task T) *OpencodeAgent[S, K, T] {
	slog.Info("initializing opencode client")

	baseURL := opts.BaseUrl
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &OpencodeAgent[S, K, T]{
		httpClient:  httpClient,
		baseURL:     baseURL,
		sessionMode: normalizeSessionMode(opts.SessionMode),
		spec:        &task,
	}
}

func normalizeSessionMode(mode core.TaskAgentSessionMode) core.TaskAgentSessionMode {
	switch mode {
	case "", core.TaskAgentSessionModeNewPerMessage:
		return core.TaskAgentSessionModeNewPerMessage
	case core.TaskAgentSessionModeReusePerClient:
		return core.TaskAgentSessionModeReusePerClient
	default:
		slog.Warn("unknown opencode session mode, defaulting to new session per message", "mode", mode)
		return core.TaskAgentSessionModeNewPerMessage
	}
}

func (c *OpencodeAgent[S, K, T]) SetAgent(agent string) {
	c.optionsMu.Lock()
	defer c.optionsMu.Unlock()

	c.agent = &agent
}

func (c *OpencodeAgent[S, K, T]) ClearAgent() {
	c.optionsMu.Lock()
	defer c.optionsMu.Unlock()

	c.agent = nil
}

func (c *OpencodeAgent[S, K, T]) SetModel(providerID, modelID string) {
	c.optionsMu.Lock()
	defer c.optionsMu.Unlock()

	c.model = &modelConfig{
		providerID: providerID,
		modelID:    modelID,
	}
}

func (c *OpencodeAgent[S, K, T]) ClearModel() {
	c.optionsMu.Lock()
	defer c.optionsMu.Unlock()

	c.model = nil
}

func (c *OpencodeAgent[S, K, T]) promptOptions() (*string, *modelConfig) {
	c.optionsMu.RLock()
	defer c.optionsMu.RUnlock()

	var agent *string
	if c.agent != nil {
		agentValue := *c.agent
		agent = &agentValue
	}

	var model *modelConfig
	if c.model != nil {
		model = &modelConfig{
			providerID: c.model.providerID,
			modelID:    c.model.modelID,
		}
	}

	return agent, model
}
