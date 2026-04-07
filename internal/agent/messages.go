package agent

import (
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
