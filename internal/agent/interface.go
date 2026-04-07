package agent

import "net/http"

type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

type ClientMessage struct {
	Text string
}

type Client[T Actionable] interface {
	CheckHealth() (HealthStatus, error)
	Prompt(message ClientMessage) (PromptResult[T], error)
}

type ClientOptions struct {
	BaseUrl    string
	HTTPClient *http.Client
}

type openCodeClient[T Actionable] struct {
	httpClient     *http.Client
	baseURL        string
	requestCounter int
}
