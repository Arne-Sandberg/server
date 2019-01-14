package manager

import (
	"runtime"
	"time"

	"github.com/freecloudio/freecloud/models"
)

type StatsManager struct {
	version   string
	startTime time.Time
}

var statsManager *StatsManager

// CreateStatsManager creates a new singleton StatsManager which can be used immediately
func CreateStatsManager(version string) *StatsManager {
	if statsManager != nil {
		return statsManager
	}

	statsManager = &StatsManager{
		version:   version,
		startTime: time.Now(),
	}
	return statsManager
}

// GetStatsManager returns the singleton instance of the StatsManager
func GetStatsManager() *StatsManager {
	return statsManager
}

func (mgr *StatsManager) GetSystemStats() *models.SystemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	uptime := time.Since(mgr.startTime)
	sessionCount, err := GetAuthManager().GetSessionCount()
	if err != nil {
		sessionCount = -1
	}

	return &models.SystemStats{
		Version:       mgr.version,
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
