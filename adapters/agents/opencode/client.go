package opencode

import (
	"github.com/Timkarx/diane/core"
	"log/slog"
	"net/http"
)

func (c *OpencodeAgent[S, K, T]) CheckHealth() (core.HealthStatus, error) {
	slog.Info("req /global/health")

	var health core.HealthStatus
	if err := c.doJSON(http.MethodGet, "/global/health", nil, &health); err != nil {
		return core.HealthStatus{}, err
	}

	return health, nil
}

func (c *OpencodeAgent[S, K, T]) ScheduleTask(message S) (K, error) {
	res, err := c.prompt(message.ToText())
	if err != nil {
		slog.Error("schedule task failed", "error", err)
		var zero K
		return zero, err
	}

	structured, err := res.Structured()
	if err != nil {
		slog.Error("decode structured output failed", "error", err)
		var zero K
		return zero, err
	}

	if c.spec != nil {
		if err := (*c.spec).ExecuteEffect(message, structured); err != nil {
			return structured, err
		}
	}

	return structured, nil
}
