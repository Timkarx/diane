package agent

import "net/http"

type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

type PromptResult struct {
	Info  AssistantMessage `json:"info"`
	Parts []Part           `json:"parts"`
}

type ClientMessage struct {
	Text   string
	Format ResponseFormat
}

type Client interface {
	CheckHealth() (HealthStatus, error)
	Prompt(message ClientMessage) (PromptResult, error)
}

type ClientOptions struct {
	BaseUrl    string
	HTTPClient *http.Client
}

type openCodeClient struct {
	httpClient     *http.Client
	baseURL        string
	requestCounter int
}
