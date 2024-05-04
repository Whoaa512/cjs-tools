package collections

func Map[T, U any](f func(T) U, s []T) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
