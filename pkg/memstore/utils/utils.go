package utils

func ClonePtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	return new(*v)
}
