package golangperf

import "testing"

var sSumByte byte
var sSumInt int
var sSumRune rune

// var sTestS = "hellosjfoeijfoiewjofjewofewhuuuuuuuuihiuhiglugluglbhftydtrtrxkhvljhljblibbhcghch"

var sTestS = "龘靐齉齾龖鱻麤饢爨癵籱驫灪纞虋讟靏鬱钃癴齼灥畾齉"

/*
理论上，迭代utf8会慢一些，但是要注意，数据类型的处理，会影响性能。
比如，求和的时候，反而更慢，我怀疑是类型转换带来的消耗，抵消了不用解码utf8的性能提升。
byte求取会比int类型求和更慢。
*/

// go test -bench Benchmark_RangeString -benchmem
func Benchmark_RangeString(b *testing.B) {
	bytesContent := []byte(sTestS)
	b.Run("RangeBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, ch := range bytesContent {
				sSumByte += ch
			}
		}
	})

	s1 := sTestS
	b.Run("RangeUtf8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, ch := range s1 {
				sSumRune += ch
			}
		}
	})
}

func Benchmark_ByteVsInt(b *testing.B) {
	const count = 1000
	b.Run("AddByte", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < count; j++ {
				sSumByte += byte(i)
			}
		}
	})
	b.Run("AddInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < count; j++ {
				sSumInt += i
			}
		}
	})
}
