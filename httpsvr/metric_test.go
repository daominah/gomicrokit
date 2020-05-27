package httpsvr

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestCalcPercentile(t *testing.T) {
	a := make([]time.Duration, 0)
	for i := 1; i <= 100; i++ {
		a = append(a, time.Duration(i))
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
		reality := CalcPercentile(a, c.percentile)
		if reality != c.expect {
			t.Errorf("CalcPercentile %v: expect %v, reality %v", i, c.expect, reality)
		}
	}
	b := make([]time.Duration, 0)
	CalcPercentile(b, 0.5)
}

func TestMetric(t *testing.T) {
	m := NewMemoryMetric()
	wg := &sync.WaitGroup{}
	path0, path1 := "path0", "path1"
	a0, a1 := make([]int, 0), make([]int, 0)
	for i := 1; i <= 1000; i++ {
		a0 = append(a0, i)
		a1 = append(a1, 2000+2*i)
	}
	rand.Shuffle(1000, func(i int, j int) {
		a0[i], a0[j] = a0[j], a0[i]
	})
	rand.Shuffle(1000, func(i int, j int) {
		a1[i], a1[j] = a1[j], a1[i]
	})
	for i := 0; i < len(a0); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Add(-1)
			m.Count(path0, 1)
			m.Duration(path0, time.Duration(a0[i]))
			m.Count(path1, 1)
			m.Duration(path1, time.Duration(a1[i]))
		}(i)
	}
	wg.Wait()
	a := m.GetCurrentMetric()
	if len(a) != 2 {
		t.Fatal(len(a))
	}
	if a[0].Path != path0 || a[1].Path != path1 {
		t.Error(a)
	}
	if a[0].Percentile68 != 683 || a[0].Percentile95 != 955 || a[0].Percentile997 != 998 {
		t.Error(a[0])
	}
	//if a[1].Percentile68 != 0 || a[1].Percentile95 != 0 || a[1].Percentile997 != 0 {
	//	t.Error(a[1])
	//}
}
