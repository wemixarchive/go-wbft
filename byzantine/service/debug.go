package service

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type CallMetrics struct {
	mu         sync.Mutex
	callCounts map[string]int64 // caller -> count
	lastCalls  map[string]time.Time
}

type CallStat struct {
	Count    int64
	LastCall time.Time
}

func NewCallMetrics() *CallMetrics {
	return &CallMetrics{
		callCounts: make(map[string]int64),
		lastCalls:  make(map[string]time.Time),
	}
}

func (cm *CallMetrics) GetStats() map[string]CallStat {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	stats := make(map[string]CallStat)
	for caller, count := range cm.callCounts {
		stats[caller] = CallStat{
			Count:    count,
			LastCall: cm.lastCalls[caller],
		}
	}
	return stats
}

func (cm *CallMetrics) LogAndReset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for caller, count := range cm.callCounts {
		if count > 0 {
			log.Trace("[byzantine] Call metrics",
				"caller", caller,
				"count", count,
				"last_call", cm.lastCalls[caller])
		}
	}
}

func (cm *CallMetrics) TrackCall(caller string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.callCounts[caller]++
	cm.lastCalls[caller] = time.Now()
}

func GetCallerInfo() string {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	parts := strings.Split(file, "/")
	fileName := parts[len(parts)-1]

	return fmt.Sprintf("%s:%d %s", fileName, line, fn.Name())
}

// GetGoroutineID get Goroutine ID (for debug)
func GetGoroutineID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine"))[0]
	return idField
}
