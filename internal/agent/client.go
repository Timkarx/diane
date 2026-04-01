package agent

import (
	"io"
	"net/http"
	"fmt"
)

func (c *openCodeClient) CheckHealth() (string, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/global/health", c.baseUrl))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func NewOpenCodeClient(opts ClientOptions) *openCodeClient {
	baseUrl := opts.BaseUrl
	if baseUrl == "" {
		baseUrl = "http://localhost:4096"
	}
	return &openCodeClient{
		httpClient: http.Client{},
		baseUrl: baseUrl,
	}
}
