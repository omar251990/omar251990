package monitor

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// SystemMonitor monitors system resources
type SystemMonitor struct {
	mu            sync.RWMutex
	cpuUsage      float64
	memoryUsage   float64
	lastCPUStat   cpuStat
	lastCheckTime time.Time
}

type cpuStat struct {
	user   uint64
	nice   uint64
	system uint64
	idle   uint64
	iowait uint64
	irq    uint64
	softirq uint64
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor() *SystemMonitor {
	sm := &SystemMonitor{
		lastCheckTime: time.Now(),
	}

	// Start monitoring loop
	go sm.monitorLoop()

	return sm
}

// monitorLoop runs periodic system checks
func (sm *SystemMonitor) monitorLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sm.updateCPUUsage()
		sm.updateMemoryUsage()
	}
}

// GetCPUUsage returns current CPU usage percentage
func (sm *SystemMonitor) GetCPUUsage() float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.cpuUsage
}

// GetMemoryUsage returns current memory usage percentage
func (sm *SystemMonitor) GetMemoryUsage() float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.memoryUsage
}

// GetDiskUsage returns disk usage percentage
func (sm *SystemMonitor) GetDiskUsage() (float64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0, err
	}

	// Calculate disk usage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	usagePercent := float64(used) / float64(total) * 100.0
	return usagePercent, nil
}

// GetNetworkStats returns network statistics
func (sm *SystemMonitor) GetNetworkStats() (map[string]interface{}, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := make(map[string]interface{})
	scanner := bufio.NewScanner(file)

	// Skip first two lines (headers)
	scanner.Scan()
	scanner.Scan()

	var totalRxBytes, totalTxBytes uint64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)

		if len(fields) < 10 {
			continue
		}

		// Interface name
		iface := strings.TrimSuffix(fields[0], ":")

		// Skip loopback
		if iface == "lo" {
			continue
		}

		// Parse RX and TX bytes
		rxBytes, _ := strconv.ParseUint(fields[1], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[9], 10, 64)

		totalRxBytes += rxBytes
		totalTxBytes += txBytes
	}

	stats["rx_bytes"] = totalRxBytes
	stats["tx_bytes"] = totalTxBytes
	stats["rx_mb"] = float64(totalRxBytes) / 1024.0 / 1024.0
	stats["tx_mb"] = float64(totalTxBytes) / 1024.0 / 1024.0

	return stats, nil
}

// GetProcessStats returns process statistics
func (sm *SystemMonitor) GetProcessStats() (map[string]interface{}, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := map[string]interface{}{
		"goroutines":     runtime.NumGoroutine(),
		"alloc_mb":       float64(m.Alloc) / 1024.0 / 1024.0,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024.0 / 1024.0,
		"sys_mb":         float64(m.Sys) / 1024.0 / 1024.0,
		"num_gc":         m.NumGC,
		"heap_objects":   m.HeapObjects,
	}

	return stats, nil
}

// updateCPUUsage updates the CPU usage percentage
func (sm *SystemMonitor) updateCPUUsage() {
	stat, err := readCPUStat()
	if err != nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Calculate CPU usage since last check
	if sm.lastCPUStat.user != 0 {
		totalDelta := float64((stat.user + stat.nice + stat.system + stat.idle + stat.iowait + stat.irq + stat.softirq) -
			(sm.lastCPUStat.user + sm.lastCPUStat.nice + sm.lastCPUStat.system + sm.lastCPUStat.idle + sm.lastCPUStat.iowait + sm.lastCPUStat.irq + sm.lastCPUStat.softirq))

		idleDelta := float64(stat.idle - sm.lastCPUStat.idle)

		if totalDelta > 0 {
			sm.cpuUsage = (1.0 - (idleDelta / totalDelta)) * 100.0
		}
	}

	sm.lastCPUStat = stat
	sm.lastCheckTime = time.Now()
}

// updateMemoryUsage updates the memory usage percentage
func (sm *SystemMonitor) updateMemoryUsage() {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		value, _ := strconv.ParseUint(fields[1], 10, 64)

		switch fields[0] {
		case "MemTotal:":
			memTotal = value
		case "MemAvailable:":
			memAvailable = value
		}

		if memTotal > 0 && memAvailable > 0 {
			break
		}
	}

	if memTotal > 0 {
		sm.mu.Lock()
		sm.memoryUsage = float64(memTotal-memAvailable) / float64(memTotal) * 100.0
		sm.mu.Unlock()
	}
}

// readCPUStat reads CPU statistics from /proc/stat
func readCPUStat() (cpuStat, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStat{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return cpuStat{}, fmt.Errorf("failed to read /proc/stat")
	}

	line := scanner.Text()
	fields := strings.Fields(line)

	if len(fields) < 8 || fields[0] != "cpu" {
		return cpuStat{}, fmt.Errorf("invalid /proc/stat format")
	}

	stat := cpuStat{
		user:    parseUint64(fields[1]),
		nice:    parseUint64(fields[2]),
		system:  parseUint64(fields[3]),
		idle:    parseUint64(fields[4]),
		iowait:  parseUint64(fields[5]),
		irq:     parseUint64(fields[6]),
		softirq: parseUint64(fields[7]),
	}

	return stat, nil
}

func parseUint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
