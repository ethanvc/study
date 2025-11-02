package prefixmatch

import (
	"fmt"
	"math/rand"
	"testing"
)

// 生成测试数据
func generateTestData() (patterns []string, hitQueries []string, missQueries []string) {
	// 生成150个模式
	patterns = make([]string, 150)
	prefixes := []string{"user", "order", "product", "payment", "shipping", "account", "admin", "api", "data", "config"}
	actions := []string{"create", "update", "delete", "read", "list", "search", "export", "import", "sync", "verify"}
	suffixes := []string{"info", "detail", "summary", "report", "status", "history", "log", "stats", "view", "edit"}

	for i := 0; i < 150; i++ {
		prefix := prefixes[i%len(prefixes)]
		action := actions[(i/len(prefixes))%len(actions)]
		suffix := suffixes[(i/(len(prefixes)*len(actions)))%len(suffixes)]
		patterns[i] = fmt.Sprintf("%s/%s/%s/%d", prefix, action, suffix, i)
	}

	// 生成命中查询（80%）- 从模式中选择
	hitCount := 120 // 150 * 0.8
	hitQueries = make([]string, hitCount)
	for i := 0; i < hitCount; i++ {
		hitQueries[i] = patterns[rand.Intn(len(patterns))]
	}

	// 生成未命中查询（20%）
	missCount := 30 // 150 * 0.2
	missQueries = make([]string, missCount)
	notExistPrefixes := []string{"nonexist", "invalid", "missing", "unknown", "notfound"}
	for i := 0; i < missCount; i++ {
		prefix := notExistPrefixes[i%len(notExistPrefixes)]
		missQueries[i] = fmt.Sprintf("%s/%s/%d", prefix, "action", i)
	}

	return patterns, hitQueries, missQueries
}

// BenchmarkPrefixMatch_Insert 测试插入性能
func BenchmarkPrefixMatch_Insert(b *testing.B) {
	patterns, _, _ := generateTestData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := NewTree()
		for j, pattern := range patterns {
			tree.Insert(pattern, j)
		}
	}
}

// BenchmarkPrefixMatch_Search_Hit 测试命中查询性能
func BenchmarkPrefixMatch_Search_Hit(b *testing.B) {
	patterns, hitQueries, _ := generateTestData()

	// 预先构建树
	tree := NewTree()
	for i, pattern := range patterns {
		tree.Insert(pattern, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := hitQueries[i%len(hitQueries)]
		tree.Search(query)
	}
}

// BenchmarkPrefixMatch_Search_Miss 测试未命中查询性能
func BenchmarkPrefixMatch_Search_Miss(b *testing.B) {
	patterns, _, missQueries := generateTestData()

	// 预先构建树
	tree := NewTree()
	for i, pattern := range patterns {
		tree.Insert(pattern, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := missQueries[i%len(missQueries)]
		tree.Search(query)
	}
}

// BenchmarkPrefixMatch_Search_Mixed 测试混合查询性能（80%命中 + 20%未命中）
func BenchmarkPrefixMatch_Search_Mixed(b *testing.B) {
	patterns, hitQueries, missQueries := generateTestData()

	// 预先构建树
	tree := NewTree()
	for i, pattern := range patterns {
		tree.Insert(pattern, i)
	}

	// 混合查询列表：80%命中 + 20%未命中
	totalQueries := len(hitQueries) + len(missQueries)
	mixedQueries := make([]string, totalQueries)
	copy(mixedQueries[:len(hitQueries)], hitQueries)
	copy(mixedQueries[len(hitQueries):], missQueries)

	// 打乱顺序
	rand.Shuffle(len(mixedQueries), func(i, j int) {
		mixedQueries[i], mixedQueries[j] = mixedQueries[j], mixedQueries[i]
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := mixedQueries[i%len(mixedQueries)]
		tree.Search(query)
	}
}

// BenchmarkPrefixMatch_Sequential 测试顺序操作（插入后立即查询）
func BenchmarkPrefixMatch_Sequential(b *testing.B) {
	patterns, hitQueries, _ := generateTestData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := NewTree()

		// 插入
		for j, pattern := range patterns {
			tree.Insert(pattern, j)
		}

		// 查询
		for _, query := range hitQueries {
			tree.Search(query)
		}
	}
}

// BenchmarkPrefixMatch_Parallel_Search 测试并行查询性能
func BenchmarkPrefixMatch_Parallel_Search(b *testing.B) {
	patterns, hitQueries, missQueries := generateTestData()

	// 预先构建树
	tree := NewTree()
	for i, pattern := range patterns {
		tree.Insert(pattern, i)
	}

	// 混合查询
	mixedQueries := append(hitQueries, missQueries...)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			query := mixedQueries[i%len(mixedQueries)]
			tree.Search(query)
			i++
		}
	})
}

// BenchmarkPrefixMatch_LongPatterns 测试长模式匹配
func BenchmarkPrefixMatch_LongPatterns(b *testing.B) {
	// 生成更长的模式
	patterns := make([]string, 150)
	for i := 0; i < 150; i++ {
		patterns[i] = fmt.Sprintf("very/long/path/with/many/segments/segment%d/subsegment/detail/info/%d", i%10, i)
	}

	tree := NewTree()
	for i, pattern := range patterns {
		tree.Insert(pattern, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		tree.Search(pattern)
	}
}

// BenchmarkPrefixMatch_ShortPatterns 测试短模式匹配
func BenchmarkPrefixMatch_ShortPatterns(b *testing.B) {
	// 生成短模式
	patterns := make([]string, 150)
	for i := 0; i < 150; i++ {
		patterns[i] = fmt.Sprintf("p%d", i)
	}

	tree := NewTree()
	for i, pattern := range patterns {
		tree.Insert(pattern, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		tree.Search(pattern)
	}
}

// BenchmarkPrefixMatch_CommonPrefix 测试具有公共前缀的模式
func BenchmarkPrefixMatch_CommonPrefix(b *testing.B) {
	// 所有模式都有相同的前缀
	patterns := make([]string, 150)
	commonPrefix := "common/prefix/path"
	for i := 0; i < 150; i++ {
		patterns[i] = fmt.Sprintf("%s/branch%d/leaf%d", commonPrefix, i%10, i)
	}

	tree := NewTree()
	for i, pattern := range patterns {
		tree.Insert(pattern, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		tree.Search(pattern)
	}
}

// 验证测试数据生成的正确性
func TestGenerateTestData(t *testing.T) {
	patterns, hitQueries, missQueries := generateTestData()

	// 验证模式数量
	if len(patterns) != 150 {
		t.Errorf("Expected 150 patterns, got %d", len(patterns))
	}

	// 验证命中查询数量（80%）
	if len(hitQueries) != 120 {
		t.Errorf("Expected 120 hit queries (80%%), got %d", len(hitQueries))
	}

	// 验证未命中查询数量（20%）
	if len(missQueries) != 30 {
		t.Errorf("Expected 30 miss queries (20%%), got %d", len(missQueries))
	}

	// 构建树并验证命中率
	tree := NewTree()
	for i, pattern := range patterns {
		err := tree.Insert(pattern, i)
		if err != nil {
			t.Errorf("Insert pattern %s failed: %v", pattern, err)
		}
	}

	// 验证所有命中查询确实能找到
	hitCount := 0
	for _, query := range hitQueries {
		if node := tree.Search(query); node != nil {
			hitCount++
		}
	}
	if hitCount != len(hitQueries) {
		t.Errorf("Expected all %d hit queries to succeed, only %d succeeded", len(hitQueries), hitCount)
	}

	// 验证所有未命中查询确实找不到
	missCount := 0
	for _, query := range missQueries {
		if node := tree.Search(query); node == nil {
			missCount++
		}
	}
	if missCount != len(missQueries) {
		t.Errorf("Expected all %d miss queries to fail, only %d failed", len(missQueries), missCount)
	}

	t.Logf("Successfully generated test data:")
	t.Logf("  Patterns: %d", len(patterns))
	t.Logf("  Hit queries: %d (%.1f%%)", len(hitQueries), float64(len(hitQueries))/float64(len(patterns))*100)
	t.Logf("  Miss queries: %d (%.1f%%)", len(missQueries), float64(len(missQueries))/float64(len(patterns))*100)
	t.Logf("  Actual hit rate: %.1f%%", float64(hitCount)/float64(len(hitQueries))*100)
	t.Logf("  Actual miss rate: %.1f%%", float64(missCount)/float64(len(missQueries))*100)
}
