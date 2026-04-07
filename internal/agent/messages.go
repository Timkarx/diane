package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
)

var ErrNoStructuredOutput = errors.New("response did not include structured output")

type PromptResult[T Actionable] struct {
	Info  AssistantMessage `json:"info"`
	Parts []Part           `json:"parts"`

	structured    T
	hasStructured bool
}

func (p *PromptResult[T]) UnmarshalJSON(data []byte) error {
	var payload struct {
		Info  json.RawMessage `json:"info"`
		Parts []Part          `json:"parts"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if len(payload.Info) > 0 {
		if err := json.Unmarshal(payload.Info, &p.Info); err != nil {
			return fmt.Errorf("decode response info: %w", err)
		}

		var structured struct {
			Structured json.RawMessage `json:"structured"`
		}
		if err := json.Unmarshal(payload.Info, &structured); err != nil {
			return fmt.Errorf("decode structured output metadata: %w", err)
		}

		if len(structured.Structured) > 0 && string(structured.Structured) != "null" {
			if err := json.Unmarshal(structured.Structured, &p.structured); err != nil {
				return fmt.Errorf("decode structured output: %w", err)
			}
			p.hasStructured = true
			p.Info.Structured = p.structured
		}
	}

	p.Parts = payload.Parts
	return nil
}

func (p PromptResult[T]) Structured() (T, error) {
	if !p.hasStructured {
		var zero T
		return zero, ErrNoStructuredOutput
	}

	return p.structured, nil
}

func (p PromptResult[T]) AsPlainText() []string {
	var text []string
	for _, part := range p.Parts {
		var meta struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(part.union, &meta); err != nil {
			slog.Error("decoding response part failed", "error", err)
			continue
		}

		switch meta.Type {
		case "text":
			textPart, err := part.AsTextPart()
			if err != nil {
				slog.Error("decode text part failed", "error", err)
				continue
			}
			text = append(text, textPart.Text)
		}
	}
	return text
}

func (p PromptResult[T]) DebugPrint() {
	info, err := json.MarshalIndent(p.Info, "", "  ")
	if err != nil {
		slog.Error("marshal response info failed", "error", err)
	} else {
		fmt.Printf("info:\n%s\n", info)
	}

	for i, part := range p.Parts {
		encoded, err := json.MarshalIndent(part, "", "  ")
		if err != nil {
			slog.Error("marshal response part failed", "index", i, "error", err)
			continue
		}

		fmt.Printf("part %d:\n%s\n", i, encoded)
	}
}

func AnalyzeApartementListingPrompt(listing string) ClientMessage {
	return ClientMessage{
		Text: listing,
	}
}
