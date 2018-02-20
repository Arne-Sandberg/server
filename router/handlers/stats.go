package handlers

import (
	"runtime"
	"time"

	"github.com/freecloudio/freecloud/models/api"
	macaron "gopkg.in/macaron.v1"
)

func (s Server) StatsHandler(c *macaron.Context) {

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.Data["response"] = api.StatsResponse{
		Success:    true,
		Version:    "TBD",
		Uptime:     1 * time.Second,
		Goroutines: runtime.NumGoroutine(),
		Memory: api.MemoryStats{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			System:     m.Sys,
			NumGC:      m.NumGC,
		},
	}
}
