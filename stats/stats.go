package stats

import (
	"time"
	"runtime"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"
	"github.com/freecloudio/freecloud/auth"
)

var (
	Version   string
	StartTime time.Time
)

func Init(version string, startTime time.Time) {
	Version = version
	StartTime = startTime
}

func GetSystemStats() (*models.SystemStats){
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	uptime := time.Since(StartTime)

	return &models.SystemStats{
		Version: Version,
		AllocMem: m.Alloc,
		TotalAllocMem: m.TotalAlloc,
		SystemMem: m.Sys,
		NumGC: m.NumGC,
		GoVersion: runtime.Version(),
		NumGoroutines: uint32(runtime.NumGoroutine()),
		NumSessions: auth.TotalSessionCount(),
		Uptime: &models.Duration{ Seconds: uptime.Seconds(), Nanos: int32(uptime.Nanoseconds()) },
	}
}