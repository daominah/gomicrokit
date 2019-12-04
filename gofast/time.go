package gofast

import (
	"time"

	"github.com/daominah/gomicrokit/log"
)

// return location +07:00
func VietnamTimeLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Infof("cannot load vietnam time location vietnam", err)
	}
	return loc
}

// return now in location +07:00
func VietnamTimeNow() time.Time {
	return time.Now().In(VietnamTimeLoc())
}

// return ex: 2006-01-02T15:04:05+07:00
func VietnamTimeNowIso() string {
	return VietnamTimeNow().Format(time.RFC3339)
}