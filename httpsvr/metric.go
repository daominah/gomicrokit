package httpsvr

import (
	"container/heap"
	"sync"
	"time"

	"fmt"

	"github.com/daominah/gomicrokit/gofast"
)

// Metric monitors number of requests, duration of requests
type Metric interface {
	// Count increases value of arg path by arg count,
	// the count value usually is 1
	Count(path string, count int)
	// Count increases value of arg path by arg dur, and push the dur value to
	// a heap (for calculating percentile)
	Duration(path string, dur time.Duration)
	// Reset set all count and duration of requests to 0
	Reset()

	// returns an array of MetricRowString
	GetPrevMetric() []string
	// returns an array of MetricRowString
	GetCurrentMetric() []string
	// percentile is in [0, 1]
	GetDurationPercentile(path string, percentile float64) time.Duration
}

type MemoryMetric struct {
	prev    map[string]*MetricRow
	current map[string]*MetricRow
	mutex   *sync.Mutex
}

// NewMemoryMetric return a memory implement of Metric interface
func NewMemoryMetric() *MemoryMetric {
	ret := &MemoryMetric{
		prev:    make(map[string]*MetricRow),
		current: make(map[string]*MetricRow),
		mutex:   &sync.Mutex{},
	}
	gofast.NewCron(ret.Reset, 24*time.Hour, 17*time.Hour)
	return ret
}

func (m *MemoryMetric) Count(path string, count int) {
	m.mutex.Lock()
	if _, found := m.current[path]; !found {
		m.current[path] = NewMetricRow()
	}
	m.current[path].Count += count
	m.mutex.Unlock()
}

func (m *MemoryMetric) Duration(path string, dur time.Duration) {
	m.mutex.Lock()
	if _, found := m.current[path]; !found {
		m.current[path] = NewMetricRow()
	}
	m.current[path].TotalDuration += dur
	heap.Push(m.current[path], dur)
	m.mutex.Unlock()
}

func (m *MemoryMetric) Reset() {
	m.mutex.Lock()
	m.prev = m.current
	m.current = make(map[string]*MetricRow)
	m.mutex.Unlock()
}

func (m *MemoryMetric) GetPrevMetric() []string {
	// TODO: write GetPrevMetric
	return nil
}

func (m *MemoryMetric) GetCurrentMetric() []string {
	// TODO: write GetCurrentMetric
	ret := make([]string, 0)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for path, row := range m.current {
		ret = append(ret, fmt.Sprintf("path: %v, metric: %v", path, row.String()))
	}
	return ret
}

func (m *MemoryMetric) GetDurationPercentile(path string, percentile float64) time.Duration {
	// TODO: write GetPercentile
	return 0
}

func (h *MetricRow) Len() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return len(*h.Durations)
}
func (h *MetricRow) Less(i, j int) bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return (*h.Durations)[i] < (*h.Durations)[j]
}
func (h *MetricRow) Swap(i, j int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	(*(h.Durations))[i], (*(h.Durations))[j] = (*(h.Durations))[j], (*(h.Durations))[i]
}

func (h *MetricRow) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length
	*(h.Durations) = append(*(h.Durations), x.(time.Duration))
}

func (h *MetricRow) Pop() interface{} {
	old := *(h.Durations)
	n := len(old)
	x := old[n-1]
	*(h.Durations) = old[0 : n-1]
	return x
}

type MetricRow struct {
	Count         int
	TotalDuration time.Duration
	Durations     *[]time.Duration
	mutex         *sync.Mutex // for Durations
}

func NewMetricRow() *MetricRow {
	durs := make([]time.Duration, 0)
	return &MetricRow{
		Durations: &durs,
		mutex:     &sync.Mutex{}}
}

func (h MetricRow) String() string {
	return fmt.Sprintf("count: %v, dur: %v, percentile: %v",
		h.Count, h.TotalDuration, h.Durations)
}
