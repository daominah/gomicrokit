package maths

import (
	"testing"
)

func TestIndexInts(t *testing.T) {
	if IndexInts([]int{1, 6, 8, 9, 3, 7}, 8) != 2 {
		t.Error()
	}
	if IndexInts([]int{1, 6, 8, 9, 3, 7}, 10) != -1 {
		t.Error()
	}
}

func TestIndexStrings(t *testing.T) {
	if IndexStrings([]string{"1", "6", "8", "9", "3", "7"}, "8") != 2 {
		t.Error()
	}
	if IndexStrings([]string{"1", "6", "8", "9", "3", "7"}, "10") != -1 {
		t.Error()
	}
}
