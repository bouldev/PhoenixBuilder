package utils

func InsertHead[T any](elem0 T, l []T) []T {
	o := make([]T, len(l)+1)
	o[0] = elem0
	for i, e := range l {
		o[i+1] = e
	}
	return o
}
