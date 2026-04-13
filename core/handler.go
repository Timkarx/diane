package core

type Callback[T Actionable] func(T)

func ExecuteHandler[T Actionable](r PromptResult[T], callback Callback[T]) error {
	action, err := r.Structured()
	if err != nil {
		return err
	}
	if err := action.Validate(); err != nil {
		return err
	}
	if !action.ShouldAct() {
		return nil
	}
	callback(action)
	return nil
}
