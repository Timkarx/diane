package agent

import (
	"fmt"
	"io"
	"net/http"
	"log/slog"
)

func (c *openCodeClient) CheckHealth() (string, error) {
	slog.Info("Req: /check_health")
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

func (c *openCodeClient) EvaluateInput(input string) {
	sessionId := c.createSession()
	c.sendMessage(sessionId, input)
}

func NewOpenCodeClient(opts ClientOptions) *openCodeClient {
	slog.Info("Initializing new OpenCode client")
	baseUrl := opts.BaseUrl
	if baseUrl == "" {
		baseUrl = "http://localhost:4096"
	}
	return &openCodeClient{
		httpClient: http.Client{},
		baseUrl:    baseUrl,
	}
}
