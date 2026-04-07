package agent

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewOutputFormatUsesActionSchema(t *testing.T) {
	format, err := newOutputFormat[ListingDecision]()
	if err != nil {
		t.Fatalf("newOutputFormat() error = %v", err)
	}
	if format == nil {
		t.Fatal("newOutputFormat() returned nil format")
	}

	got, err := format.AsOutputFormatJsonSchema()
	if err != nil {
		t.Fatalf("AsOutputFormatJsonSchema() error = %v", err)
	}

	if got.Type != "json_schema" {
		t.Fatalf("Type = %q, want %q", got.Type, "json_schema")
	}

	encodedGot, err := json.Marshal(got.Schema)
	if err != nil {
		t.Fatalf("json.Marshal(got.Schema) error = %v", err)
	}

	encodedWant, err := json.Marshal(ListingDecision{}.Schema())
	if err != nil {
		t.Fatalf("json.Marshal(want.Schema) error = %v", err)
	}

	if string(encodedGot) != string(encodedWant) {
		t.Fatalf("Schema = %s, want %s", encodedGot, encodedWant)
	}
}

func TestPromptResultStructuredDecodesAction(t *testing.T) {
	raw := []byte(`{"info":{"structured":{"should_notify":true,"summary":"fits"}},"parts":[]}`)

	var result PromptResult[ListingDecision]
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	got, err := result.Structured()
	if err != nil {
		t.Fatalf("Structured() error = %v", err)
	}

	want := ListingDecision{
		ShouldNotify: true,
		Summary:      "fits",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Structured() = %#v, want %#v", got, want)
	}
}

func TestExecuteHandlerUsesActionTrigger(t *testing.T) {
	raw := []byte(`{"info":{"structured":{"should_notify":true,"summary":"fits"}},"parts":[]}`)

	var result PromptResult[ListingDecision]
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	called := false
	err := ExecuteHandler(result, func(decision ListingDecision) {
		called = true
		if decision.Summary != "fits" {
			t.Fatalf("Summary = %q, want %q", decision.Summary, "fits")
		}
	})
	if err != nil {
		t.Fatalf("ExecuteHandler() error = %v", err)
	}
	if !called {
		t.Fatal("ExecuteHandler() did not invoke callback")
	}

	raw = []byte(`{"info":{"structured":{"should_notify":false,"summary":"skip"}},"parts":[]}`)
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	called = false
	err = ExecuteHandler(result, func(ListingDecision) {
		called = true
	})
	if err != nil {
		t.Fatalf("ExecuteHandler() error = %v", err)
	}
	if called {
		t.Fatal("ExecuteHandler() invoked callback for non-actionable result")
	}
}
