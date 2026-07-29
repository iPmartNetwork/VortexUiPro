// Package sys provides system monitoring utilities (memory, CPU, connections).
package sys

import (
	"os"
	"sync"

	"github.com/shirou/gopsutil/v4/process"
)

var (
	selfProc     *process.Process
	selfProcOnce sync.Once
)

// SelfRSS returns the resident set size of the current process in bytes.
// Unlike runtime.MemStats.Sys (a never-shrinking high-water mark), RSS reflects
// current physical memory usage. Returns 0 when unavailable.
func SelfRSS() uint64 {
	selfProcOnce.Do(func() {
		if p, err := process.NewProcess(int32(os.Getpid())); err == nil {
			selfProc = p
		}
	})
	if selfProc == nil {
		return 0
	}
	if mi, err := selfProc.MemoryInfo(); err == nil && mi != nil {
		return mi.RSS
	}
	return 0
}

// CPUPercent returns the CPU usage percentage of the current process.
func CPUPercent() float64 {
	selfProcOnce.Do(func() {
		if p, err := process.NewProcess(int32(os.Getpid())); err == nil {
			selfProc = p
		}
	})
	if selfProc == nil {
		return 0
	}
	pct, err := selfProc.CPUPercent()
	if err != nil {
		return 0
	}
	return pct
}

// NumFDs returns the number of open file descriptors (handles on Windows) for the process.
func NumFDs() int {
	selfProcOnce.Do(func() {
		if p, err := process.NewProcess(int32(os.Getpid())); err == nil {
			selfProc = p
		}
	})
	if selfProc == nil {
		return 0
	}
	fds, err := selfProc.NumFDs()
	if err != nil {
		return 0
	}
	return int(fds)
}

// NumThreads returns the number of threads used by the process.
func NumThreads() int32 {
	selfProcOnce.Do(func() {
		if p, err := process.NewProcess(int32(os.Getpid())); err == nil {
			selfProc = p
		}
	})
	if selfProc == nil {
		return 0
	}
	n, err := selfProc.NumThreads()
	if err != nil {
		return 0
	}
	return n
}
