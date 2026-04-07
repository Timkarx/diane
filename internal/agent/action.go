package agent

type Actionable interface {
	Schema() JSONSchema
	ShouldAct() bool
	Validate() error
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
