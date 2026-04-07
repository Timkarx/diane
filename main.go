package main

import (
	"diane/internal/agent"
	"diane/internal/telegram_bot"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"runtime"
	"strings"
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
	client := agent.NewOpenCodeClient(clientOpts)

	res, err := client.Prompt("What gurantees does c++ give about function arg evaluation?")
	if err != nil {
		os.Exit(1)
	}
	_ , err = bot.SendMessage(strings.Join(res.AsPlainText(), " "))
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
