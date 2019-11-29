package gofast

import (
	"errors"
	"fmt"
	"reflect"

	errors2 "github.com/pkg/errors"
)

var (
	ErrNonPointerOutput = errors.New("non pointer output param")
	ErrSourceNil        = errors.New("source is nil")
	ErrUnexpected       = errors.New("unexpected error feelsbadman")
)

// copy same fields of 2 struct,
// destination d must be a pointer, source s can be pointer or value
func CopySameFields(d interface{}, s interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors2.Wrap(err, fmt.Sprintf("%v", r))
		}
	}()
	if reflect.ValueOf(d).Kind() != reflect.Ptr {
		return ErrNonPointerOutput
	}
	dType, dValue := reflect.TypeOf(d).Elem(), reflect.ValueOf(d).Elem()
	sType, sValue := reflect.TypeOf(s), reflect.ValueOf(s)
	if sValue.Kind() == reflect.Ptr {
		sValue = sValue.Elem()
	}
	if !sValue.IsValid() {
		return ErrSourceNil
	}
	for i := 0; i < dType.NumField(); i++ {
		dField := dType.FieldByIndex([]int{i})
		if sValue.FieldByName(dField.Name).IsValid() {
			sField, found := sType.FieldByName(dField.Name)
			if !found || sField.Type.Kind() != dField.Type.Kind() {
				continue
			}
			dValue.FieldByName(dField.Name).Set(sValue.FieldByName(dField.Name))
		}
	}
	return nil
}

func checkNilInterface(x interface{}, isDebug bool) (result bool) {
	defer func() {
		r := recover()
		if r != nil {
			result = false
		}
		if isDebug {
			fmt.Printf(
				"result: %-5v, x==nil: %-5v, panic: %-5v, x: %v(%T)\n",
				result, x == nil, r != nil, x, x)
		}
	}()
	if x == nil {
		// only untyped nil (ex: nil error) return here
		return true
	}
	if reflect.ValueOf(x).IsNil() {
		// panic if x is not chan, func, interface, map, pointer, or slice
		return true
	}
	return false
}

// Problem: `x == nil` still returns true if x is a nil struct pointer,
// checkNilInterface will return false in above case,
// more examples are in `reflect_test.go`
func CheckNilInterface(x interface{}) (result bool) {
	return checkNilInterface(x, false)
}
