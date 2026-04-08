package agent

import (
	"fmt"
	"net/url"
	"strings"
)

type Actionable interface {
	Schema() JSONSchema
	ShouldAct() bool
	Validate() error
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

type ListingNotification struct {
	Text   string
	Link   string
	Photos []string
}

type Unstructured struct{}

func (Unstructured) Schema() JSONSchema {
	return nil
}

func (Unstructured) ShouldAct() bool {
	return false
}

func (Unstructured) Validate() error {
	return nil
}

type ListingDecision struct {
	ShouldNotify bool   `json:"should_notify"`
	Summary      string `json:"summary,omitempty"`
}

func (ListingDecision) Schema() JSONSchema {
	return JSONSchema{
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

func (d ListingDecision) ToNotification(input ListingInput) ListingNotification {
	return ListingNotification{
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
