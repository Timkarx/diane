package agent

import (
	"log/slog"
)

type Callback func()

func ExecuteHandler(r PromptResult, callback func(ListingDecision)) error {
	decision, err := DecodeStructured[ListingDecision](r)
	if err != nil {
		return err
	}
	if err := decision.Validate(); err != nil {
		return err
	}
	if !decision.ShouldNotify {
		return nil
	}
	callback(decision)
	return nil
}
