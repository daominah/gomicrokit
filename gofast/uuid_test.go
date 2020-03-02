package gofast

import (
	"testing"
)

func TestGenUUID(t *testing.T) {
	id := GenUUID()
	if len(id) != 32 {
		t.Error()
	}
}

func TestGenUUID2(t *testing.T) {
	ids := make(map[string]bool)
	n := 100000
	for i := 0; i < n; i++ {
		ids[GenUUID()] = true
	}
	if len(ids) != n {
		t.Error(len(ids))
	}
}

func TestGenUUID3(t *testing.T) {
	id := GenUUIDWithHyphen()
	if len(id) != 36 {
		t.Error()
	}
}

func TestGenUUID4(t *testing.T) {
	ids := make(map[string]bool)
	n := 100000
	for i := 0; i < n; i++ {
		ids[GenUUIDWithHyphen()] = true
	}
	if len(ids) != n {
		t.Error(len(ids))
	}
}
