package copy

import (
	"fmt"
	"testing"
)

func BenchmarkOptions_NumberOfWorkers_0(b *testing.B) {
	var num int64 = 0 // 0 or 1 = single-threaded
	opt := Options{NumberOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}

func BenchmarkOptions_NumberOfWorkers_2(b *testing.B) {
	var num int64 = 2
	opt := Options{NumberOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}

func BenchmarkOptions_NumberOfWorkers_4(b *testing.B) {
	var num int64 = 4
	opt := Options{NumberOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}

func BenchmarkOptions_NumberOfWorkers_8(b *testing.B) {
	var num int64 = 8
	opt := Options{NumberOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}
