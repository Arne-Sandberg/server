package stats

import (
	"runtime"
	"time"

	"github.com/freecloudio/freecloud/manager"
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
	sessionCount, err := manager.GetAuthManager().GetSessionCount()
	if err != nil {
		sessionCount = -1
	}

	return &models.SystemStats{
		Version:       Version,
		AllocMem:      int64(m.Alloc),
		TotalAllocMem: int64(m.TotalAlloc),
		SystemMem:     int64(m.Sys),
		NumGC:         int64(m.NumGC),
		GoVersion:     runtime.Version(),
		NumGoroutines: int64(runtime.NumGoroutine()),
		NumSessions:   int64(sessionCount),
		Uptime:        int64(uptime.Seconds()),
	}
}
