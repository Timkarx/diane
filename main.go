package main

import (
	"diane/internal/agent"
	"diane/internal/telegram_bot"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"runtime"
	"log"
)

func setup() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	cores := runtime.GOMAXPROCS(0)
	fmt.Println("Number of cores: ", cores)
}

func main() {
	setup()
	bot_token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chat_id := os.Getenv("TELEGRAM_CHAT_ID")

	_ = telegram_bot.NewTelegramBot(bot_token, chat_id)

	clientOpts := agent.ClientOptions{}
	client := agent.NewOpenCodeClient(clientOpts)

	prompt_bytes, err := os.ReadFile("test/bad_prompt.md")
	if err != nil {
		log.Fatal(err)
	}

	msg := agent.ClientMessage{
		Text: string(prompt_bytes),
		Format: agent.JSONSchemaFormat{
			Schema: agent.JSONSchema{
				"type": "object",
				"properties": map[string]any{
					"should_notify": map[string]any{
						"type":        "boolean",
						"desctiption": "true if this item fits all the specified criteria",
					},
					"summary": map[string]any{
						"type":        "string",
						"description": "Short summary of the listing, if applicable",
					},
				},
			},
		},
	}

	res, err := client.Prompt(msg)
	if err != nil {
		os.Exit(1)
	}
	res.DebugPrint()
	os.Exit(0)
}
