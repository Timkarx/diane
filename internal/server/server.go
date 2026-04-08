package server

import (
	"diane/internal/agent"
	"diane/internal/telegram_bot"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func New() http.Handler {
	mux := http.NewServeMux()
	client := agent.NewOpenCodeClient[agent.ListingDecision](agent.ClientOptions{})
	bot := telegram_bot.NewTelegramBot(os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID"))

	instruction_bytes, err := os.ReadFile("test/instructions.md")
	if err != nil {
		log.Fatal(err)
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"healthy": true,
		})
	})

	mux.HandleFunc("POST /listing-summary", func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract request body
		payload, err := parseRequest(r)
		if err != nil {
			writeJSON(w, 422, map[string]any{
				"error": "Invalid payload shape",
			})
			return
		}

		go func(text string) {
			response, err := client.Prompt(agent.AnalyzeApartementListingPrompt(text, string(instruction_bytes)))
			if err != nil {
				log.Printf("listing summary prompt failed: %v", err)
				return
			}

			callback := func(actionable agent.ListingDecision) {
				if err := bot.SendMessage(actionable.Summarize()); err != nil {
					log.Printf("listing summary notification failed: %v", err)
				}
			}

			if err := agent.ExecuteHandler(response, callback); err != nil {
				log.Printf("listing summary handler failed: %v", err)
			}
		}(payload.Text)

		// 2. Acknowledge the request and let the workflow finish asynchronously.
		writeJSON(w, http.StatusAccepted, map[string]any{
			"status": "processing",
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
