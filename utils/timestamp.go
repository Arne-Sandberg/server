package utils

import (
	"time"
)

func GetTimestampNow() int64 {
	return time.Now().UTC().Unix()
}