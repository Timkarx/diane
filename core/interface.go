package core

import "net/http"

type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

type JSONSchema map[string]interface{}

type TaskSpec[K any] interface {
	Schema() JSONSchema
	ExecuteEffect(K) error
	Validate() error
}

type TaskAgentMessage struct {
	Text string
}

type TaskAgentOptions struct {
	BaseUrl    string
	HTTPClient *http.Client
}

type TaskAgent[K any, T TaskSpec[K]] interface {
	ScheduleTask(TaskAgentMessage) (K, error)
}
