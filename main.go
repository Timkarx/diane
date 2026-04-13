package main

import (
	"diane/adapters/opencode"
	"diane/core"
	"diane/internal/telegram_bot"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"runtime"
)

type TaskStructuredOoutput struct {
	ShouldNotify bool   `json:"should_notify"`
	Summary      string `json:"summary,omitempty"`
}

type ListingAnalysisTask struct{
	bot *telegram_bot.TelegramBot
}

func (ListingAnalysisTask) Schema() core.JSONSchema {
	return core.JSONSchema{
		"type": "object",
		"properties": map[string]any{
			"should_notify": map[string]any{
				"type":        "boolean",
				"description": "true if this item fits all the specified criteria",
			},
			"summary": map[string]any{
				"type":        "string",
				"description": "Short summary of the listing, if applicable",
			},
		},
		"required":             []string{"should_notify"},
		"additionalProperties": false,
	}
}

func (ListingAnalysisTask) ExecuteEffect(o TaskStructuredOoutput) error {

}

func (ListingAnalysisTask) Validate() error {
	return nil
}

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

	bot := telegram_bot.NewTelegramBot(os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID"))

	clientOpts := core.TaskAgentOptions{}
	client := opencode.NewOpenCodeClient[TaskStructuredOoutput, ListingAnalysisTask](clientOpts)

	instruction_bytes, err := os.ReadFile("test/instructions.md")
	if err != nil {
		log.Fatal(err)
	}
	prompt_bytes, err := os.ReadFile("test/good_prompt.md")
	if err != nil {
		log.Fatal(err)
	}

	msg := core.AnalyzeApartementListingPrompt(string(prompt_bytes), string(instruction_bytes))
	res, err := client.Prompt(msg)
	if err != nil {
		os.Exit(1)
	}

	callback := func(actionable core.ListingDecision) {
		notification := actionable.ToNotification(core.ListingInput{})
		if err := bot.SendMessage(notification); err != nil {
			log.Printf("telegram notification failed: %v", err)
		}
	}

	core.ExecuteHandler(res, callback)

	os.Exit(0)
}
