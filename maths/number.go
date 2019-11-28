package maths

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
	minInt  = -maxInt - 1
)

func MinInts(ints ...int) int {
	r := maxInt
	for _, i := range ints {
		if r > i {
			r = i
		}
	}
	return r
}

func MaxInts(ints ...int) int {
	r := minInt
	for _, i := range ints {
		if r < i {
			r = i
		}
	}
	return r
}
