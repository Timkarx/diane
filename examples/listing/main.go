package main

import (
	"fmt"
	"github.com/Timkarx/diane/adapters/agents/opencode"
	"github.com/Timkarx/diane/adapters/notifications/telegram"
	"github.com/Timkarx/diane/core"
	"github.com/joho/godotenv"
	"log"
	"os"
	"runtime"
)

type AgentInput struct {
	Listing string   `json:"listing"`
	Link    string   `json:"link"`
	Photos  []string `json:"photos,omitempty"`
}

func (a AgentInput) ToText() string {
	return a.Listing
}

type TaskStructuredOoutput struct {
	ShouldNotify bool   `json:"should_notify"`
	Summary      string `json:"summary,omitempty"`
}

type ListingAnalysisTask struct {
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

func (l *ListingAnalysisTask) ExecuteEffect(input AgentInput, out TaskStructuredOoutput) error {
	if out.ShouldNotify {
		notification := telegram_bot.TelegramMessage{
			Text:   out.Summary,
			Link:   input.Link,
			Photos: input.Photos,
		}
		if err := l.bot.SendMessage(notification); err != nil {
			return err
		}
	}
	return nil
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

	clientOpts := core.TaskAgentOptions{
		SessionMode: core.TaskAgentSessionModeReusePerClient,
	}
	task := ListingAnalysisTask{&bot}
	client := opencode.NewOpenCodeClient[AgentInput, TaskStructuredOoutput, *ListingAnalysisTask](clientOpts, &task)

	instruction_bytes, err := os.ReadFile("test/instructions.md")
	if err != nil {
		log.Fatal(err)
	}
	prompt_bytes, err := os.ReadFile("test/good_prompt.md")
	if err != nil {
		log.Fatal(err)
	}

	msg := AgentInput{
		Listing: string(instruction_bytes) + "/n" + string(prompt_bytes),
		Link:    "http://example.org",
		Photos:  []string{},
	}
	client.ScheduleTask(msg)

	os.Exit(0)
}
