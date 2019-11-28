package maths

import (
	"testing"
)

func TestMinInts(t *testing.T) {
	for i, c := range []struct {
		ints []int
		min  int
	}{
		{[]int{8, 7}, 7},
		{[]int{7, 8}, 7},
		{[]int{5, 1, 6, 3, 8, 7}, 1},
		{[]int{2, 4, 9}, 2},
		{nil, maxInt},
	} {
		if MinInts(c.ints...) != c.min {
			t.Error(i, MinInts(c.ints...), c.min)
		}
	}
}

func TestMaxInts(t *testing.T) {
	for i, c := range []struct {
		ints []int
		max  int
	}{
		{[]int{8, 7}, 8},
		{[]int{7, 8}, 8},
		{[]int{5, 1, 6, 3, 8, 7}, 8},
		{[]int{2, 4, 9}, 9},
		{nil, minInt},
	} {
		if MaxInts(c.ints...) != c.max {
			t.Error(i, MaxInts(c.ints...), c.max)
		}
	}
}
