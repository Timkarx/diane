package core

import "net/http"

type TaskAgentSessionMode string

const (
	TaskAgentSessionModeNewPerMessage  TaskAgentSessionMode = "new_per_message"
	TaskAgentSessionModeReusePerClient TaskAgentSessionMode = "reuse_per_client"
)

type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

type JSONSchema map[string]interface{}

type TaskSpec[S TaskAgentMessage, K any] interface {
	Schema() JSONSchema
	ExecuteEffect(S, K) error
	Validate() error
}

type TaskAgentMessage interface {
	ToText() string
}

type TaskAgentOptions struct {
	BaseUrl     string
	Port        int
	HTTPClient  *http.Client
	SessionMode TaskAgentSessionMode
}

type TaskAgent[K any, S TaskAgentMessage, T TaskSpec[S, K]] interface {
	ScheduleTask(TaskAgentMessage) (K, error)
}
