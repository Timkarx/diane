package core

type Callback[T TaskSpec] func(T)

func ExecuteHandler[T TaskSpec](r TaskResult[T], callback Callback[T]) error {
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
