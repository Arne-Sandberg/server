package manager

import (
	"runtime"
	"time"

	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/repository"
)

// SystemManager provides
type SystemManager struct {
	version     string
	startTime   time.Time
	sessionRep  *repository.SessionRepository
	fileInfoRep *repository.FileInfoRepository
}

var systemManager *SystemManager

// CreateSystemManager creates a new singleton StatsManager which can be used immediately
func CreateSystemManager(version string, sessionRep *repository.SessionRepository, fileInfoRep *repository.FileInfoRepository) *SystemManager {
	if systemManager != nil {
		return systemManager
	}

	systemManager = &SystemManager{
		version:     version,
		startTime:   time.Now(),
		sessionRep:  sessionRep,
		fileInfoRep: fileInfoRep,
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
	sessionCount, err := mgr.sessionRep.Count()
	if err != nil {
		sessionCount = -1
	}
	fileInfoCount, err := mgr.fileInfoRep.Count()
	if err != nil {
		fileInfoCount = -1
	}

	stats = &models.SystemStats{
		Version:       mgr.version,
		AllocMem:      int64(m.Alloc),
		TotalAllocMem: int64(m.TotalAlloc),
		SystemMem:     int64(m.Sys),
		NumGC:         int64(m.NumGC),
		GoVersion:     runtime.Version(),
		NumGoroutines: int64(runtime.NumGoroutine()),
		NumSessions:   sessionCount,
		NumFileInfos:  fileInfoCount,
		Uptime:        int64(uptime.Seconds()),
	}
	return
}
