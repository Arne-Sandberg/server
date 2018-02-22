package stats

import (
	"time"
)

var (
	Version   string
	StartTime time.Time
)

func Init(version string, startTime time.Time) {
	Version = version
	StartTime = startTime
}
