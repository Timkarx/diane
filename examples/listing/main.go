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
	"strings"
	"net/url"
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

type ListingInput struct {
	Listing string   `json:"listing"`
	Link    string   `json:"link"`
	Photos  []string `json:"photos,omitempty"`
}

func (i ListingInput) Validate() error {
	if strings.TrimSpace(i.Listing) == "" {
		return fmt.Errorf("listing is required")
	}
	if err := validateURL("link", i.Link); err != nil {
		return err
	}
	for index, photo := range i.Photos {
		if err := validateURL(fmt.Sprintf("photos[%d]", index), photo); err != nil {
			return err
		}
	}
	return nil
}

type ListingDecision struct {
	ShouldNotify bool   `json:"should_notify"`
	Summary      string `json:"summary,omitempty"`
}

func (ListingDecision) Schema() core.JSONSchema {
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

func (d ListingDecision) ShouldAct() bool {
	return d.ShouldNotify
}

func (ListingDecision) Validate() error {
	return nil
}

func (d ListingDecision) ToNotification(input ListingInput) core.ListingNotification {
	return core.ListingNotification{
		Text:   d.Summary,
		Link:   input.Link,
		Photos: input.Photos,
	}
}

func validateURL(field string, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("%s is required", field)
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("%s must be a valid URL", field)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use http or https", field)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s must include a host", field)
	}

	return nil
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
