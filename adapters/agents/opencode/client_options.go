package opencode

import (
	"github.com/Timkarx/diane/core"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

const defaultBaseURL = "http://localhost"
const defaultPort = 4096

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

func NewOpenCodeClient[S core.TaskAgentMessage, K any, T core.TaskSpec[S, K]](opts core.TaskAgentOptions, task T) *OpencodeAgent[S, K, T] {
	slog.Info("initializing opencode client")

	baseURL := opts.BaseUrl
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	resolvedBaseURL, err := resolveBaseURL(baseURL, opts.Port)
	if err != nil {
		slog.Warn("invalid opencode base url, defaulting to localhost", "base_url", baseURL, "port", opts.Port, "error", err)
		resolvedBaseURL, _ = resolveBaseURL(defaultBaseURL, defaultPort)
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &OpencodeAgent[S, K, T]{
		httpClient:  httpClient,
		baseURL:     resolvedBaseURL,
		sessionMode: normalizeSessionMode(opts.SessionMode),
		spec:        &task,
	}
}

func resolveBaseURL(baseURL string, port int) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if parsed.Scheme == "" {
		parsed, err = url.Parse("http://" + baseURL)
		if err != nil {
			return "", err
		}
	}

	resolvedPort := port
	if resolvedPort == 0 {
		if parsed.Port() == "" {
			resolvedPort = defaultPort
		}
	}

	if resolvedPort != 0 {
		host := parsed.Hostname()
		if host == "" {
			host = parsed.Host
		}
		parsed.Host = net.JoinHostPort(host, strconv.Itoa(resolvedPort))
	}

	return parsed.String(), nil
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
