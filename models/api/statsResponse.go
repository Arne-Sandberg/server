package api

import "time"

type StatsResponse struct {
	Success    bool          `json:"success,omitempty"`
	Version    string        `json:"version,omitempty"`
	Uptime     time.Duration `json:"uptime,omitempty"`
	Memory     MemoryStats   `json:"memory,omitempty"`
	Goroutines int           `json:"goroutines"`
	Sessions   int           `json:"sessions"`
}

type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"totalAlloc"`
	System     uint64 `json:"system"`
	NumGC      uint32 `json:"numGC"`
}
