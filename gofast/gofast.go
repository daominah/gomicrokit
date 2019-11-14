package gofast

import (
	"errors"
	"reflect"
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
			err = ErrUnexpected
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
