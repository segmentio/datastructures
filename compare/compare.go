package compare

import "constraints"

// Function is a comparison function for ordered types.
func Function[T constraints.Ordered](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return +1
	default:
		return 0
	}
}
