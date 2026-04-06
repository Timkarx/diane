package server

import (
	"diane/internal/agent"
	"encoding/json"
	"fmt"
	"net/http"
)

func New() http.Handler {
	mux := http.NewServeMux()
	clientOpts := agent.ClientOptions{}
	client := agent.NewOpenCodeClient(clientOpts)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"healthy": true,
		})
	})

	mux.HandleFunc("POST /prompt", func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract request body
		payload, err := parseRequest(r)
		if err != nil {
			writeJSON(w, 422, map[string]any{
				"error": "Invalid payload shape",
			})
			return
		}

		// 2. Pass it to the client
		response, err := client.Prompt(payload.Text)
		if err != nil {
			writeJSON(w, 500, map[string]any{
				"error": "Internal error",
			})
			return
		}

		// 3. Return client response
		writeJSON(w, 200, map[string]any{
			"text": response.AsPlainText(),
		})

	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseRequest(r *http.Request) (Payload, error) {
	var prompt Payload
	err := json.NewDecoder(r.Body).Decode(&prompt)
	defer r.Body.Close()
	if err != nil {
		return Payload{}, fmt.Errorf("Invalid payload")
	}
	return prompt, nil
}
