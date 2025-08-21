package service

import (
	"fmt"
	"runtime"
	"strconv"
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
			log.Trace("BYZ: Call metrics",
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

const (
	functionSplit = "."
	traceSplit    = "/"
)

func FTrace(fname string, fCount int) string {
	var fn string
	fn = fTrace(2)
	for i := 1; i < fCount; i++ {
		temp := fTrace(2 + i)
		if len(temp) == 0 {
			break
		}
		fn = temp + "\n" + fn
	}
	fn = fname + " Function Trace\n" + fn + "\n"
	return fn
}

func fTrace(depth int) string {
	functionSplit := func(function string) string {
		functions := strings.Split(function, functionSplit)
		return functions[len(functions)-1]
	}

	fileSplit := func(file string) []string {
		return strings.Split(file, traceSplit)
	}

	// pc, file, line, ok := runtime.Caller(depth)
	pc, file, line, ok := runtime.Caller(depth)
	files := fileSplit(file)

	if ok {
		f := runtime.FuncForPC(pc)
		function := functionSplit(f.Name())
		// return f.Name() + "()" + ":" + file + "/" + linestr
		return "[" + files[len(files)-2] + "/" + files[len(files)-1] + ", " + function + "():" + strconv.Itoa(line) + "]"
	}

	return ""
}
