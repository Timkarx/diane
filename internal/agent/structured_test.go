package agent

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJSONSchemaFormatMarshalJSON(t *testing.T) {
	schema := JSONSchema{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{"type": "string"},
		},
	}

	encoded, err := JSONSchemaFormat{
		Schema:     schema,
		RetryCount: 2,
	}.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var got struct {
		RetryCount int        `json:"retryCount"`
		Schema     JSONSchema `json:"schema"`
		Type       string     `json:"type"`
	}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got.Type != "json_schema" {
		t.Fatalf("Type = %q, want %q", got.Type, "json_schema")
	}

	if got.RetryCount != 2 {
		t.Fatalf("RetryCount = %d, want %d", got.RetryCount, 2)
	}

	if !reflect.DeepEqual(got.Schema, schema) {
		t.Fatalf("Schema = %#v, want %#v", got.Schema, schema)
	}
}

func TestDecodeStructured(t *testing.T) {
	result := PromptResult{
		Info: AssistantMessage{
			Structured: map[string]any{
				"title": "Bright studio",
				"price": 1200,
			},
		},
	}

	type listing struct {
		Title string `json:"title"`
		Price int    `json:"price"`
	}

	got, err := DecodeStructured[listing](result)
	if err != nil {
		t.Fatalf("DecodeStructured() error = %v", err)
	}

	want := listing{
		Title: "Bright studio",
		Price: 1200,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("listing = %#v, want %#v", got, want)
	}
}
