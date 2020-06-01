package metric

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestCalcPercentile(t *testing.T) {
	a := *NewMetricRow()
	for i := 1; i <= 100; i++ {
		a.Durations.InsertNoReplace(Duration(i))
	}
	for i, c := range []struct {
		percentile float64
		expect     time.Duration
	}{
		{0.6827, 69},
		{0.9545, 96},
		{0.9973, 100},
		{0.99, 99},
		{68, 100},
	} {
		reality := calcRowPercentile(a, c.percentile)
		if reality != c.expect {
			t.Errorf("CalcPercentile %v: expect %v, reality %v", i, c.expect, reality)
		}
	}
	// try empty row
	if calcRowPercentile(*NewMetricRow(), 0.5) != 0 {
		t.Error("try empty row")
	}
}

func TestMetric(t *testing.T) {
	m := NewMemoryMetric()
	wg := &sync.WaitGroup{}
	path0, path1 := "path0", "path1"
	a0, a1 := make([]int, 0), make([]int, 0)
	for i := 1; i <= 1000; i++ {
		a0 = append(a0, i)
		a0 = append(a0, i)
		a1 = append(a1, 2000+2*i)
	}
	rand.Shuffle(len(a0), func(i int, j int) {
		a0[i], a0[j] = a0[j], a0[i]
	})
	rand.Shuffle(len(a1), func(i int, j int) {
		a1[i], a1[j] = a1[j], a1[i]
	})
	for _, c := range []struct {
		path  string
		array []int
	}{{path0, a0}, {path1, a1}} {
		for i := 0; i < len(c.array); i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Add(-1)
				m.Count(c.path)
				m.Duration(c.path, time.Duration(c.array[i]))
			}(i)
		}
		wg.Wait()
	}
	a := m.GetCurrentMetric()
	if len(a) != 2 {
		t.Fatal(len(a))
	}
	if a[0].Key != path0 || a[1].Key != path1 {
		t.Error(a)
	}
	t.Log(a[0])
	if a[0].Percentile68 < 683 ||
		a[0].Percentile95 < 955 ||
		a[0].Percentile997 < 998 {
		t.Error()
	}
}
