package core

import (
	"encoding/json"
	"fmt"
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
func New(handlers ...Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"healthy": true,
		})
	})

	for _, h := range handlers {
		mux.HandleFunc(h.Method, h.HandlerFunc)
	}

	return mux
}

func parseRequest(r *http.Request) (ListingInput, error) {
	var input ListingInput
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return ListingInput{}, fmt.Errorf("Invalid payload")
	}
	if err := input.Validate(); err != nil {
		return ListingInput{}, err
	}
	return input, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
