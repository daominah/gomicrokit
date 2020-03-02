package gofast

// IndexInts returns the index of the first instance of sub in slice,
// or -1 if sub is not present in slice
func IndexInts(slice []int, sub int) int {
	for i, v := range slice {
		if v == sub {
			return i
		}
	}
	return -1
}

// IndexStrings returns the index of the first instance of sub in slice,
// or -1 if sub is not present in slice
func IndexStrings(slice []string, sub string) int {
	for i, v := range slice {
		if v == sub {
			return i
		}
	}
	return -1
}
