package compare

// Ordered is a type constraint enumerating primitive types that support the
// "<" and ">" operators.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~string
}

// Function is a comparison function for ordered types.
func Function[T Ordered](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return +1
	default:
		return 0
	}
}
