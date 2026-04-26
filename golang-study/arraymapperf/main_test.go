package arraymapperf

import (
	"fmt"
	"testing"
)

// BenchmarkMap_Get 测试不同数据量下 map 和 array 的查找性能
func BenchmarkMap_Get(b *testing.B) {
	sizes := []int{0, 1, 8, 15, 30, 80, 120, 255}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			m := make(map[byte]int)
			// 填充指定数量的数据
			for i := 0; i < size; i++ {
				m[byte(i)] = i * 10
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := byte(i % 256)
				_ = m[key]
			}
		})
	}
}

// BenchmarkArray_Get 测试不同数据量下数组的查找性能
func BenchmarkArray_Get(b *testing.B) {
	sizes := []int{0, 1, 8, 15, 30, 80, 120, 255}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			var arr [256]*int
			// 填充指定数量的数据
			for i := 0; i < size; i++ {
				val := i * 10
				arr[i] = &val
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := byte(i % 256)
				_ = arr[key]
			}
		})
	}
}

// BenchmarkLinear_Get 测试不同数据量下线性搜索的查找性能
func BenchmarkLinear_Get(b *testing.B) {
	sizes := []int{0, 1, 8, 15, 24, 30, 80, 120, 255}

	type KeyValue struct {
		key   byte
		value int
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// 使用切片存储键值对
			slice := make([]KeyValue, size)
			for i := 0; i < size; i++ {
				slice[i] = KeyValue{key: byte(i), value: i * 10}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := byte(i % 256)
				// 线性搜索
				var found bool
				for j := 0; j < len(slice); j++ {
					if slice[j].key == key {
						_ = slice[j].value
						found = true
						break
					}
				}
				_ = found
			}
		})
	}
}
