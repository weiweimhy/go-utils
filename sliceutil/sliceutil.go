package sliceutil

// Chunk splits values into chunks of size n. It returns nil when n <= 0.
func Chunk[T any](values []T, n int) [][]T {
	if n <= 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(values)+n-1)/n)
	for len(values) > 0 {
		end := n
		if end > len(values) {
			end = len(values)
		}
		chunks = append(chunks, values[:end])
		values = values[end:]
	}
	return chunks
}

// Unique returns values with duplicates removed while preserving first-seen order.
func Unique[T comparable](values []T) []T {
	seen := make(map[T]struct{}, len(values))
	out := make([]T, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// Map transforms values with fn.
func Map[T, U any](values []T, fn func(T) U) []U {
	out := make([]U, 0, len(values))
	for _, value := range values {
		out = append(out, fn(value))
	}
	return out
}

// Filter returns values for which fn returns true.
func Filter[T any](values []T, fn func(T) bool) []T {
	out := make([]T, 0, len(values))
	for _, value := range values {
		if fn(value) {
			out = append(out, value)
		}
	}
	return out
}
