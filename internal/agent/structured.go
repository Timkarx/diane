package agent

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrNoStructuredOutput = errors.New("response did not include structured output")

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

func (p PromptResult) StructuredJSON() (json.RawMessage, error) {
	if p.Info.Structured == nil {
		return nil, ErrNoStructuredOutput
	}

	encoded, err := json.Marshal(p.Info.Structured)
	if err != nil {
		return nil, fmt.Errorf("marshal structured output: %w", err)
	}

	return encoded, nil
}

func (p PromptResult) DecodeStructured(dst any) error {
	encoded, err := p.StructuredJSON()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(encoded, dst); err != nil {
		return fmt.Errorf("decode structured output: %w", err)
	}

	return nil
}

func DecodeStructured[T any](result PromptResult) (T, error) {
	var value T
	if err := result.DecodeStructured(&value); err != nil {
		return value, err
	}

	return value, nil
}
