package gofast

import (
	"testing"
)

func TestCopySameFields(t *testing.T) {
	type S struct {
		F1 int
		F2 string
		F3 bool
		F5 string
	}
	type D1 struct {
		F1 int
		F2 string
	}
	type D2 struct {
		F1 string
		F2 map[string]string
		F3 bool
		F4 int
	}

	// fail
	s := S{F1: 1, F2: "f2", F3: true, F5: "f5"}
	var d1 D1
	err := CopySameFields(d1, s)
	if err != errNonPointerOutput {
		t.Error(err)
	}
	err = CopySameFields(&d1, nil)
	if err != errSourceNil {
		t.Error(err)
	}

	// success
	err = CopySameFields(&d1, s)
	if err != nil {
		t.Error(err)
	}
	expectD1 := D1{F1: s.F1, F2: s.F2}
	if d1 != expectD1 {
		t.Error(d1, expectD1)
	}

	err = CopySameFields(&d1, &s)
	if err != nil {
		t.Error(err)
	}
	if d1 != expectD1 {
		t.Error(d1, expectD1)
	}

	var d2 D2
	err = CopySameFields(&d2, s)
	if err != nil {
		t.Error(err)
	}
	if d2.F1 != "" || d2.F2 != nil || d2.F3 != s.F3 || d2.F4 != 0 {
		t.Error()
	}
}

func TestCheckNilInterface(t *testing.T) {
	type MyStruct struct{ Name string }

	var nilStructPtr *MyStruct
	myStruct := MyStruct{Name: "Name0"}
	var nilErr error
	var nilStringPtr *string
	var str string = "pussy"
	var nilSclice []int
	initedSclice := []int{1, 2}

	for i, c := range []struct {
		x      interface{}
		expect bool
	}{
		// interface{}(nilStructPtr) == nil will return false
		{nilStructPtr, true},
		{myStruct, false},
		{nilErr, true},
		{nilStringPtr, true},
		{str, false},
		{nilSclice, true},
		{initedSclice, false},
	} {
		if isNil := checkNilInterface(c.x, false); isNil != c.expect {
			t.Error(i, c.x, isNil)
		}
	}
}
