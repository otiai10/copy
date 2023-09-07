package copy

import (
	"fmt"
	"testing"
)

func BenchmarkOptions_NumOfWorkers_0(b *testing.B) {
	var num int64 = 0 // 0 or 1 = single-threaded
	opt := Options{NumOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}

func BenchmarkOptions_NumOfWorkers_2(b *testing.B) {
	var num int64 = 2
	opt := Options{NumOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}

func BenchmarkOptions_NumOfWorkers_4(b *testing.B) {
	var num int64 = 4
	opt := Options{NumOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}

func BenchmarkOptions_NumOfWorkers_8(b *testing.B) {
	var num int64 = 8
	opt := Options{NumOfWorkers: num}
	for i := 0; i < b.N; i++ {
		Copy("test/data/case19", fmt.Sprintf("test/data.copy/case19-%d-%d", num, i), opt)
	}
}
