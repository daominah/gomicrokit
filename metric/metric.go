package metric

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/daominah/GoLLRB/llrb"
	"github.com/daominah/gomicrokit/gofast"
)

// Metric monitors number of requests, duration of requests
type Metric interface {
	// Count increases count value of the key by 1
	Count(key string)
	// Count increases total duration of the key by dur,
	// the dur will be saved in a order statistic tree
	Duration(key string, dur time.Duration)
	// Reset set all count and duration of requests to 0,
	// In a database implement, you can persist the prevMetric
	Reset()

	// returns an array of MetricRowString
	GetCurrentMetric() []RowDisplay
	// percentile is in [0, 1]
	GetDurationPercentile(key string, percentile float64) time.Duration
	// returns an array of MetricRowString
	GetPrevMetric() []RowDisplay
}

// RowDisplay is human readable metric data of a key
type RowDisplay struct {
	// example of Key: http path_method
	Key             string
	Count           int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	Percentile68    time.Duration
	Percentile95    time.Duration
	Percentile997   time.Duration
}

// MemoryMetric implements Metric interface
type MemoryMetric struct {
	current map[string]*Row
	prev    map[string]*Row
	*sync.Mutex
}

// Row is a memory representation of RowDisplay
type Row struct {
	Count         int
	TotalDuration time.Duration
	Durations     *llrb.LLRB
	*sync.Mutex
}

// NewMemoryMetric return a memory implement of Metric interface,
// this struct's methods is safe for concurrent calls
func NewMemoryMetric() *MemoryMetric {
	ret := &MemoryMetric{
		prev:    make(map[string]*Row),
		current: make(map[string]*Row),
		Mutex:   &sync.Mutex{},
	}
	gofast.NewCron(ret.Reset, 24*time.Hour, 17*time.Hour)
	return ret
}

func (m *MemoryMetric) getRow(key string) *Row {
	m.Lock()
	row, found := m.current[key]
	if !found {
		m.current[key] = NewMetricRow()
		row = m.current[key]
	}
	m.Unlock()
	return row
}

func (m *MemoryMetric) Count(key string) {
	row := m.getRow(key)
	row.Lock()
	row.Count += 1
	row.Unlock()
}

func (m *MemoryMetric) Duration(key string, dur time.Duration) {
	row := m.getRow(key)
	row.Lock()
	row.TotalDuration += dur
	row.Durations.InsertNoReplace(Duration(dur))
	row.Unlock()
}

func (m *MemoryMetric) Reset() {
	m.Lock()
	m.prev = m.current
	m.current = make(map[string]*Row)
	m.Unlock()
}

func (m *MemoryMetric) GetCurrentMetric() []RowDisplay {
	ret := make([]RowDisplay, 0)
	m.Lock()
	for key, row := range m.current {
		ret = append(ret, row.Display(key))
	}
	m.Unlock()
	sort.Sort(SortByKey(ret))
	return ret
}
func (m *MemoryMetric) GetPrevMetric() []RowDisplay {
	ret := make([]RowDisplay, 0)
	m.Lock()
	for key, row := range m.prev {
		ret = append(ret, row.Display(key))
	}
	m.Unlock()
	sort.Sort(SortByKey(ret))
	return ret
}

func (m *MemoryMetric) GetDurationPercentile(key string, percentile float64) time.Duration {
	row := m.getRow(key)
	row.Lock()
	ret := calcRowPercentile(*row, percentile)
	row.Unlock()
	return ret
}

func NewMetricRow() *Row {
	return &Row{
		Durations: llrb.New(),
		Mutex:     &sync.Mutex{},
	}
}

func (r Row) Display(key string) RowDisplay {
	r.Lock()
	defer r.Unlock()
	ret := RowDisplay{Key: key, Count: r.Count, TotalDuration: r.TotalDuration}
	if r.Count != 0 {
		ret.AverageDuration = r.TotalDuration / time.Duration(r.Count)
	}
	ret.Percentile68 = calcRowPercentile(r, 0.6827)
	ret.Percentile95 = calcRowPercentile(r, 0.9545)
	ret.Percentile997 = calcRowPercentile(r, 0.9973)
	return ret
}

// do not lock row in this func
func calcRowPercentile(row Row, percentile float64) time.Duration {
	rank := int(math.Ceil(percentile*float64(row.Durations.Len())))
	item := row.Durations.GetByRank(rank)
	dur, ok := item.(Duration)
	if item == nil || !ok {
		return 0
	}
	return time.Duration(dur)
}

func (d RowDisplay) String() string {
	return fmt.Sprintf(
		"key: %v, count: %v, dur: %v, aveDur: %v, p68: %v, p95: %v, p99.7: %v",
		d.Key, d.Count, d.TotalDuration, d.AverageDuration,
		d.Percentile68, d.Percentile95, d.Percentile997)
}

type SortByKey []RowDisplay

func (h SortByKey) Len() int           { return len(h) }
func (h SortByKey) Less(i, j int) bool { return h[i].Key < h[j].Key }
func (h SortByKey) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// Duration is time_Duration with method Less
type Duration time.Duration

func (d Duration) Less(than llrb.Item) bool {
	tmp, _ := than.(Duration)
	return d < tmp
}
