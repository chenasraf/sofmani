package utils

func MergeMap[K comparable, V any](me map[K]V, other map[K]V) map[K]V {
	copy := map[K]V{}
	for k, v := range me {
		copy[k] = v
	}
	for k, v := range other {
		copy[k] = v
	}
	return me
}
