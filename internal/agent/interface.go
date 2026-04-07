package agent

import "net/http"

type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

type ClientMessage struct {
	Text   string
	Format ResponseFormat
}

type Client[T any] interface {
	CheckHealth() (HealthStatus, error)
	Prompt(message ClientMessage) (PromptResult[T], error)
}

type ClientOptions struct {
	BaseUrl    string
	HTTPClient *http.Client
}

type openCodeClient[T any] struct {
	httpClient     *http.Client
	baseURL        string
	requestCounter int
}
