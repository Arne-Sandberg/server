package manager

import (
	"runtime"
	"time"

	"github.com/freecloudio/server/models"
)

// SystemManager provides
type SystemManager struct {
	version   string
	startTime time.Time
}

var systemManager *SystemManager

// CreateSystemManager creates a new singleton StatsManager which can be used immediately
func CreateSystemManager(version string) *SystemManager {
	if systemManager != nil {
		return systemManager
	}

	systemManager = &SystemManager{
		version:   version,
		startTime: time.Now(),
	}
	return systemManager
}

// GetSystemManager returns the singleton instance of the StatsManager
func GetSystemManager() *SystemManager {
	return systemManager
}

// GetSystemStats returns the current stats of the system
func (mgr *SystemManager) GetSystemStats() (stats *models.SystemStats, err error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	uptime := time.Since(mgr.startTime)
	sessionCount, err := GetAuthManager().GetSessionCount()
	if err != nil {
		sessionCount = -1
	}

	stats = &models.SystemStats{
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
	return
}
