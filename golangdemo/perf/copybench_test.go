package perf

import (
	"log/slog"
	"sync"
	"testing"
)

var recordPool = sync.Pool{
	New: func() any { return &slog.Record{} },
}

//go:noinline
func consumeByValue(r slog.Record) int {
	return r.NumAttrs()
}

//go:noinline
func consumeByPointer(r *slog.Record) int {
	return r.NumAttrs()
}

func BenchmarkPassRecord(b *testing.B) {
	b.Run("byValue", func(b *testing.B) {
		b.ReportAllocs()
		var r slog.Record
		for b.Loop() {
			consumeByValue(r)
		}
	})

	b.Run("byPointer", func(b *testing.B) {
		b.ReportAllocs()
		var r slog.Record
		for b.Loop() {
			consumeByPointer(&r)
		}
	})

	b.Run("byPointer/pool", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := recordPool.Get().(*slog.Record)
			consumeByPointer(r)
			recordPool.Put(r)
		}
	})
}
