package slices

func Contains[A comparable](xs []A, y A) bool {
	for _, x := range xs {
		if x == y {
			return true
		}
	}
	return false
}
