package helpers

type Slice[T comparable] []T

func (slice Slice[T]) IndexOf(value T) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func (slice Slice[T]) Map(f func(T) T) Slice[T] {
	n := len(slice)
	a := make(Slice[T], n)

	for i, v := range slice {
		a[i] = f(v)
	}

	return a
}
