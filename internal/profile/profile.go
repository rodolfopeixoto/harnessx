// SPDX-License-Identifier: MIT

// Package profile wraps pprof for harness hot-path benchmarks.
package profile

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

type Heap struct {
	AllocBytes      uint64
	TotalAllocBytes uint64
	Sys             uint64
	NumGC           uint32
	CapturedAt      time.Time
}

func Snapshot() Heap {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Heap{
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		Sys:             m.Sys,
		NumGC:           m.NumGC,
		CapturedAt:      time.Now().UTC(),
	}
}

func WriteHeap(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	runtime.GC()
	return pprof.WriteHeapProfile(f)
}

func StartCPU(path string) (stop func() error, err error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, err
	}
	return func() error { pprof.StopCPUProfile(); return f.Close() }, nil
}

func DiffPct(before, after uint64) float64 {
	if before == 0 {
		return 0
	}
	return (float64(after) - float64(before)) / float64(before) * 100.0
}

func (h Heap) String() string {
	return fmt.Sprintf("alloc=%dB total=%dB sys=%dB gc=%d at=%s",
		h.AllocBytes, h.TotalAllocBytes, h.Sys, h.NumGC,
		h.CapturedAt.Format(time.RFC3339))
}
