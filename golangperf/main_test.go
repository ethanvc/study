package golangperf

import "testing"

var sSum int
var sTestS = "hellosjfoeijfoiewjofjewofewhuuuuuuuuihiuhiglugluglbhftydtrtrxkhvljhljblibbhcghch"

// go test -bench Benchmark_RangeString -benchmem
func Benchmark_RangeString(b *testing.B) {
	bytesContent := []byte(sTestS)
	b.Run("RangeBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, ch := range bytesContent {
				sSum += int(ch)
			}
		}
	})

	s1 := sTestS
	b.Run("RangeUtf8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, ch := range s1 {
				sSum += int(ch)
			}
		}
	})
}
