package stats

import (
	"runtime"
	"time"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
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
		AllocMem:      int64(m.Alloc),
		TotalAllocMem: int64(m.TotalAlloc),
		SystemMem:     int64(m.Sys),
		NumGC:         int64(m.NumGC),
		GoVersion:     runtime.Version(),
		NumGoroutines: int64(runtime.NumGoroutine()),
		NumSessions:   auth.TotalSessionCount(),
		Uptime:        int64(uptime.Seconds()),
	}
}
