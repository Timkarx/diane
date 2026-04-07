package agent

import (
	"encoding/json"
	"fmt"
	"errors"
	"log/slog"
)

var ErrNoStructuredOutput = errors.New("response did not include structured output")

type PromptResult[T any] struct {
	Info  AssistantMessage `json:"info"`
	Parts []Part           `json:"parts"`
}

type ResponseFormat interface {
	json.Marshaler
}

type JSONSchemaFormat struct {
	Schema     JSONSchema
	RetryCount int
}

func (f JSONSchemaFormat) MarshalJSON() ([]byte, error) {
	if len(f.Schema) == 0 {
		return nil, errors.New("json schema format requires a schema")
	}

	payload := struct {
		RetryCount *int       `json:"retryCount,omitempty"`
		Schema     JSONSchema `json:"schema"`
		Type       string     `json:"type"`
	}{
		Schema: f.Schema,
		Type:   "json_schema",
	}

	if f.RetryCount > 0 {
		payload.RetryCount = &f.RetryCount
	}

	return json.Marshal(payload)
}

func (p PromptResult[T]) StructuredJSON() (json.RawMessage, error) {
	if p.Info.Structured == nil {
		return nil, ErrNoStructuredOutput
	}

	encoded, err := json.Marshal(p.Info.Structured)
	if err != nil {
		return nil, fmt.Errorf("marshal structured output: %w", err)
	}

	return encoded, nil
}

func (p PromptResult[T]) DecodeStructured(dst any) (T, error) {
	var value T
	encoded, err := p.StructuredJSON()
	if err != nil {
		return value, err
	}

	if err := json.Unmarshal(encoded, &value); err != nil {
		return value, fmt.Errorf("decode structured output: %w", err)
	}

	return value, nil
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
		Format: JSONSchemaFormat{
			Schema: JSONSchema{
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
}
