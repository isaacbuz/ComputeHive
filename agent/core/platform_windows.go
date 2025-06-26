//go:build windows
// +build windows

package core

import (
	"fmt"
)

// getLoadAverage returns an error on Windows as load average is not available
func getLoadAverage() (map[string]float64, error) {
	return nil, fmt.Errorf("load average not available on Windows")
} 