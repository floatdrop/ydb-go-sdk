package scanner

import (
	"testing"
)

var (
	testSize = 10000
)

func BenchmarkTestScanWithColumns(b *testing.B) {
	b.ReportAllocs()
	res := PrepareScannerPerformanceTest(b.N)
	row := series{}
	res.setColumnIndexes([]string{"series_id", "title", "release_date"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for res.NextRow() {
			_ = res.Scan(&row.id, &row.title, &row.date)
		}
	}
}

func BenchmarkTestScan(b *testing.B) {
	b.ReportAllocs()
	res := PrepareScannerPerformanceTest(b.N)
	row := series{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if res.NextRow() {
			_ = res.Scan(&row.id, &row.title, &row.date)
		}
	}
}

func TestOverallApproaches(t *testing.T) {
	for k, f := range map[string]func(b *testing.B){
		"BenchmarkTestScanWithColumns": BenchmarkTestScanWithColumns,
		"BenchmarkTestScan":            BenchmarkTestScan,
	} {
		r := testing.Benchmark(f)
		t.Log(k, r.String(), r.MemString())
	}
}

func TestOverallSliceApproaches(t *testing.T) {
	sizear := []int{2, 5, 10, 20, 50, 100}
	for _, testSize = range sizear {
		t.Logf("Slice size: %d", testSize)
		for _, test := range []struct {
			name string
			f    func(b *testing.B)
		}{
			{
				"BenchmarkTestDoubleIndex",
				BenchmarkTestDoubleIndex,
			}, {
				"BenchmarkTestTempValue",
				BenchmarkTestTempValue,
			},
		} {
			r := testing.Benchmark(test.f)
			t.Log(test.name, r.String())
		}
	}
}

func BenchmarkTestSliceReduce(b *testing.B) {
	var c = make([]*column, testSize)
	for j := 0; j < testSize; j++ {
		c[j] = &column{}
	}
	b.ResetTimer()
	var row column
	for i := 0; i < b.N; i++ {
		slice := c
		for j := 0; j < testSize; j++ {
			row = *slice[0]
			slice = slice[1:]
		}
	}
	_ = row
}

func BenchmarkTestSliceIncrement(b *testing.B) {
	var slice = make([]*column, testSize)
	for j := 0; j < testSize; j++ {
		slice[j] = &column{}
	}
	cnt := 0
	var row column
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cnt = 0
		for i := 0; i < testSize; i++ {
			row = *slice[cnt]
			cnt++
		}
	}
	_ = row
}

func BenchmarkTestTempValue(b *testing.B) {
	var slice = make([]*column, testSize)
	for j := 0; j < testSize; j++ {
		slice[j] = &column{}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < testSize; i++ {
			col := slice[i]
			col.name = "test1"
			col.typeID = 1
		}
	}
}

func BenchmarkTestDoubleIndex(b *testing.B) {
	var slice = make([]*column, testSize)
	for j := 0; j < testSize; j++ {
		slice[j] = &column{}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < testSize; i++ {
			slice[i].name = "test2"
			slice[i].typeID = 1
		}
	}
}
