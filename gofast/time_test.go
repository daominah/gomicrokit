package gofast

import (
	"testing"
	"time"
)

func TestVietnamTime(t *testing.T) {
	now := VietnamTimeNow()
	nowStr := now.Format(time.RFC3339)
	if nowStr[len(nowStr)-6:] != "+07:00" {
		t.Error()
	}

}
