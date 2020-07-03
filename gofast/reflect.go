package gofast

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errNonPointerOutput = errors.New("non pointer output param")
	errSourceNil        = errors.New("source is nil")
)

// CopySameFields copies same fields of 2 struct,
// destination d must be a pointer, source s can be pointer or value.
// I am sorry, this function kill "Find Usages" and fuck up debugging, though
// it improve coding speed.
func CopySameFields(d interface{}, s interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error CopySameFields: %v", r)
		}
	}()
	if reflect.ValueOf(d).Kind() != reflect.Ptr {
		return errNonPointerOutput
	}
	dType, dValue := reflect.TypeOf(d).Elem(), reflect.ValueOf(d).Elem()
	sType, sValue := reflect.TypeOf(s), reflect.ValueOf(s)
	if sValue.Kind() == reflect.Ptr {
		sType = sType.Elem()
		sValue = sValue.Elem()
	}
	if !sValue.IsValid() {
		return errSourceNil
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

// CheckNilInterface returns true if arg x is a nil struct pointer.
// We need this func because an interface variable is nil only if both the type
// and value are nil, so expression `x == nil` will return false if x is a nil
// struct pointer.
func CheckNilInterface(x interface{}) (result bool) {
	return checkNilInterface(x, false)
}
