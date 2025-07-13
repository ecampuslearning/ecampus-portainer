package slicesx

// Iterates over elements of collection, returning an array of all elements predicate returns truthy for.
//
// Note: Unlike `FilterInPlace`, this method returns a new array.
func Filter[T any](input []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for i := range input {
		if predicate(input[i]) {
			result = append(result, input[i])
		}
	}
	return result
}

// Filter in place all elements from input that predicate returns truthy for and returns an array of the removed elements.
//
// Note: Unlike `Filter`, this method mutates input.
func FilterInPlace[T any](input []T, predicate func(T) bool) []T {
	n := 0
	for _, v := range input {
		if predicate(v) {
			input[n] = v
			n++
		}
	}
	return input[:n]
}
