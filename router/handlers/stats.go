package handlers

import (
	"runtime"
	"time"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models/api"
	"github.com/freecloudio/freecloud/stats"
	macaron "gopkg.in/macaron.v1"
)

func (s Server) StatsHandler(c *macaron.Context) {

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.Data["response"] = api.StatsResponse{
		Success:    true,
		Version:    stats.Version,
		Uptime:     time.Since(stats.StartTime),
		Goroutines: runtime.NumGoroutine(),
		GoVersion:  runtime.Version(),
		Memory: api.MemoryStats{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			System:     m.Sys,
			NumGC:      m.NumGC,
		},
		Sessions: auth.TotalSessionCount(),
	}
}
