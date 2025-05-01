package forms

type PayloadWrapper[T any] struct {
	Payload T `json:"payload"`
}

func (pw *PayloadWrapper[T]) Unwrap() T {
	return pw.Payload
}
