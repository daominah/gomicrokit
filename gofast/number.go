package gofast

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
	minInt  = -maxInt - 1
)

// MinInts find the min value in the inputs
func MinInts(ints ...int) int {
	r := maxInt
	for _, i := range ints {
		if r > i {
			r = i
		}
	}
	return r
}

// MaxInts find the max value in the inputs
func MaxInts(ints ...int) int {
	r := minInt
	for _, i := range ints {
		if r < i {
			r = i
		}
	}
	return r
}
