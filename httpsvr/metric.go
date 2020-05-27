package httpsvr

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

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
	// Reset set all count and duration of requests to 0,
	// In a database implement, you can persist the prevMetric
	Reset()

	// returns an array of MetricRowString
	GetPrevMetric() []MetricRowDisplay
	// returns an array of MetricRowString
	GetCurrentMetric() []MetricRowDisplay
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
	m.current[path].mutex.Lock()
	heap.Push(m.current[path].Durations, dur)
	m.current[path].mutex.Unlock()
	m.mutex.Unlock()
}

func (m *MemoryMetric) Reset() {
	m.mutex.Lock()
	m.prev = m.current
	m.current = make(map[string]*MetricRow)
	m.mutex.Unlock()
}

func (m *MemoryMetric) GetPrevMetric() []MetricRowDisplay {
	ret := make([]MetricRowDisplay, 0)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for path, row := range m.prev {
		ret = append(ret, row.Display(path))
	}
	sort.Sort(SortByPath(ret))
	return ret
}

func (m *MemoryMetric) GetCurrentMetric() []MetricRowDisplay {
	ret := make([]MetricRowDisplay, 0)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for path, row := range m.current {
		ret = append(ret, row.Display(path))
	}
	sort.Sort(SortByPath(ret))
	return ret
}

func (m *MemoryMetric) GetDurationPercentile(path string, percentile float64) time.Duration {
	// TODO: write GetPercentile
	return 0
}

type DurationsHeap []time.Duration

func (h DurationsHeap) Len() int           { return len(h) }
func (h DurationsHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h DurationsHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *DurationsHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length
	*h = append(*h, x.(time.Duration))
}

func (h *DurationsHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type MetricRow struct {
	Count         int
	TotalDuration time.Duration
	Durations     *DurationsHeap
	mutex         *sync.Mutex // for field Durations
}

type MetricRowDisplay struct {
	Path            string
	Count           int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	Percentile68    time.Duration
	Percentile95    time.Duration
	Percentile997   time.Duration
}

func NewMetricRow() *MetricRow {
	return &MetricRow{
		Durations: &DurationsHeap{},
		mutex:     &sync.Mutex{},
	}
}

func (h MetricRow) Display(path string) MetricRowDisplay {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	ret := MetricRowDisplay{Path: path, Count: h.Count, TotalDuration: h.TotalDuration}
	if h.Count != 0 {
		ret.AverageDuration = h.TotalDuration / time.Duration(h.Count)
	}
	ret.Percentile68 = CalcPercentile(*h.Durations, 0.6827)
	ret.Percentile95 = CalcPercentile(*h.Durations, 0.9545)
	ret.Percentile997 = CalcPercentile(*h.Durations, 0.9973)
	return ret
}

func (d MetricRowDisplay) String() string {
	return fmt.Sprintf(
		"path: %v, count: %v, dur: %v, aveDur: %v, p68: %v, p95: %v, p99.7: %v",
		d.Path, d.Count, d.TotalDuration, d.AverageDuration,
		d.Percentile68, d.Percentile95, d.Percentile997)
}

// CalcPercentile _,
// :arg percentile: in [0,1],
func CalcPercentile(sorted []time.Duration, percentile float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(percentile*float64(len(sorted)))) - 1
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}

type SortByPath []MetricRowDisplay

func (h SortByPath) Len() int           { return len(h) }
func (h SortByPath) Less(i, j int) bool { return h[i].Path < h[j].Path }
func (h SortByPath) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
