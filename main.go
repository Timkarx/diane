package main

import (
	"diane/internal/agent"
	"diane/internal/telegram_bot"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"runtime"
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

	bot := telegram_bot.NewTelegramBot(bot_token, chat_id)

	clientOpts := agent.ClientOptions{}
	client := agent.NewOpenCodeClient[agent.ListingDecision](clientOpts)

	prompt_bytes, err := os.ReadFile("test/good_prompt.md")
	if err != nil {
		log.Fatal(err)
	}

	msg := agent.AnalyzeApartementListingPrompt(string(prompt_bytes))
	res, err := client.Prompt(msg)
	if err != nil {
		os.Exit(1)
	}

	callback := func(actionable agent.ListingDecision) {
		bot.SendMessage(actionable.Summarize())
	}

	agent.ExecuteHandler(res, callback)


	os.Exit(0)
}
