package server

import (
	"diane/internal/core"
	"diane/internal/telegram_bot"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func New() http.Handler {
	mux := http.NewServeMux()
	client := core.NewOpenCodeClient[core.ListingDecision](core.ClientOptions{})
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
		input, err := parseRequest(r)
		if err != nil {
			writeJSON(w, 422, map[string]any{
				"error": "Invalid payload",
			})
			return
		}

		go func(input core.ListingInput) {
			response, err := client.Prompt(core.AnalyzeApartementListingPrompt(input.Listing, string(instruction_bytes)))
			if err != nil {
				log.Printf("listing summary prompt failed: %v", err)
				return
			}

			callback := func(actionable core.ListingDecision) {
				notification := actionable.ToNotification(input)
				if err := bot.SendMessage(notification); err != nil {
					log.Printf("listing summary notification failed: %v", err)
				}
			}

			if err := core.ExecuteHandler(response, callback); err != nil {
				log.Printf("listing summary handler failed: %v", err)
			}
		}(input)

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
