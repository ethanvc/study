package arraymapperf

func FindFromSlice[T comparable](slice []T, target T) (T, bool) {
	for _, v := range slice {
		if v == target {
			return v, true
		}
	}
	var defaultVal T
	return defaultVal, false
}

func FindFromMap[T comparable](m map[T]T, target T) (T, bool) {
	val, ok := m[target]
	return val, ok
}
