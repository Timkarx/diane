package server

import (
	"diane/core"
	"encoding/json"
	"fmt"
	"net/http"
)

func New(handlers ...core.Handler) http.Handler {
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseRequest(r *http.Request) (core.ListingInput, error) {
	var input core.ListingInput
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return core.ListingInput{}, fmt.Errorf("Invalid payload")
	}
	if err := input.Validate(); err != nil {
		return core.ListingInput{}, err
	}
	return input, nil
}
