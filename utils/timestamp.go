package utils

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

func GetTimestampNow() *timestamp.Timestamp {
	return &timestamp.Timestamp{ Seconds: time.Now().UTC().Unix(), Nanos: int32(time.Now().UTC().Nanosecond()) }
}

func GetTimestampFromTime(convTime time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{ Seconds: convTime.Unix(), Nanos: int32(convTime.Nanosecond()) }
}