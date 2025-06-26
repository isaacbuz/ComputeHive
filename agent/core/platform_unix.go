//go:build !windows
// +build !windows

package core

import (
	"github.com/shirou/gopsutil/v3/load"
)

// getLoadAverage returns system load averages for Unix-like systems
func getLoadAverage() (map[string]float64, error) {
	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}
	
	return map[string]float64{
		"1min":  avg.Load1,
		"5min":  avg.Load5,
		"15min": avg.Load15,
	}, nil
} 