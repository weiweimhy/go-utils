package maputil

import "slices"

// KeysSorted returns map keys in ascending order.
func KeysSorted[K cmpOrdered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// Clone returns a shallow copy of m.
func Clone[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	out := make(map[K]V, len(m))
	for key, value := range m {
		out[key] = value
	}
	return out
}

// Merge returns a shallow copy containing maps in order. Later maps override earlier ones.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	size := 0
	for _, m := range maps {
		size += len(m)
	}
	out := make(map[K]V, size)
	for _, m := range maps {
		for key, value := range m {
			out[key] = value
		}
	}
	return out
}

type cmpOrdered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}
