package slices

func Map[A interface{}, B interface{}](xs []A, f func(a A) B) []B {
	ys := make([]B, len(xs))
	for i, x := range xs {
		ys[i] = f(x)
	}
	return ys
}

func Contains[A comparable](xs []A, y A) bool {
	for _, x := range xs {
		if x == y {
			return true
		}
	}
	return false
}

func All[A interface{}](xs []A, f func(a A) bool) bool {
	for _, x := range xs {
		if !f(x) {
			return false
		}
	}
	return true
}

func Any[A interface{}](xs []A, f func(a A) bool) bool {
	for _, x := range xs {
		if f(x) {
			return true
		}
	}
	return false
}
