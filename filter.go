package main

func filter[T any](values []T, accept func(T) bool) []T {
	r := make([]T, 0, len(values))

	for _, v := range values {
		if accept(v) {
			r = append(r, v)
		}
	}

	return r
}

func contains[T any](values []T, accept func(T) bool) bool {
	for _, v := range values {
		if accept(v) {
			return true
		}
	}
	return false
}
