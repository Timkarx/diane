package agent

import "net/http"

type Client interface {
	CheckHealth() (string error)
}

type ClientOptions struct {
	BaseUrl string
}

type openCodeClient struct {
	httpClient http.Client
	baseUrl string
	requestCounter int
}
