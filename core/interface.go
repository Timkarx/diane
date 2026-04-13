package core

import "net/http"

type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

type TaskAgentMessage struct {
	Text string
}

type TaskAgentOptions struct {
	BaseUrl    string
	HTTPClient *http.Client
}

type TaskAgent[T Actionable] interface {
	ScheduleTask(TaskAgentMessage) PromptResult[T]
}
