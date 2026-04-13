package core

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	Method      string
	HandlerFunc func(w http.ResponseWriter, r *http.Request)
}

func HttpHandler[InputSchema any](
	method string,
	parseRequest func(*http.Request) (InputSchema, error),
	taskHandler func(InputSchema),
) Handler {
	_handler := func(w http.ResponseWriter, r *http.Request) {
		input, err := parseRequest(r)
		if err != nil {
			writeJSON(w, 422, map[string]any{
				"error": "Invalid payload",
			})
			return
		}

		go taskHandler(input)
	}

	return Handler{
		method,
		_handler,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
