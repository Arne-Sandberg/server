package stats

import (
	"math"
	"runtime"
	"time"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	"github.com/golang/protobuf/ptypes/duration"
)

var (
	Version   string
	StartTime time.Time
)

func Init(version string, startTime time.Time) {
	Version = version
	StartTime = startTime
}

func GetSystemStats() *models.SystemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	uptime := time.Since(StartTime)

	return &models.SystemStats{
		Version:       Version,
		AllocMem:      m.Alloc,
		TotalAllocMem: m.TotalAlloc,
		SystemMem:     m.Sys,
		NumGC:         m.NumGC,
		GoVersion:     runtime.Version(),
		NumGoroutines: uint32(runtime.NumGoroutine()),
		NumSessions:   auth.TotalSessionCount(),
		Uptime:        &duration.Duration{Seconds: int64(math.Round(uptime.Seconds()))},
	}
}
