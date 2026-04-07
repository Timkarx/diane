package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
)

func (p PromptResult) AsPlainText() []string {
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

func (p PromptResult) DebugPrint() {
	for i, part := range p.Parts {
		encoded, err := part.MarshalJSON()
		if err != nil {
			slog.Error("marshal response part failed", "index", i, "error", err)
			continue
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, encoded, "", "  "); err != nil {
			slog.Error("pretty print response part failed", "index", i, "error", err)
			fmt.Printf("part %d: %s\n", i, string(encoded))
			continue
		}

		fmt.Printf("part %d:\n%s\n", i, pretty.String())
	}
}
