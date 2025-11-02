# Prefix Match Benchmark 测试

## 测试规模

- **模式数量**: 150 个
- **命中查询**: 120 个 (80%)
- **未命中查询**: 30 个 (20%)

## Benchmark 测试项

### 1. BenchmarkPrefixMatch_Insert
测试插入 150 个模式的性能。

### 2. BenchmarkPrefixMatch_Search_Hit
测试命中查询的性能（所有查询都能找到结果）。

### 3. BenchmarkPrefixMatch_Search_Miss
测试未命中查询的性能（所有查询都找不到结果）。

### 4. BenchmarkPrefixMatch_Search_Mixed
测试混合查询性能（80% 命中 + 20% 未命中）。**这是最接近真实场景的测试**。

### 5. BenchmarkPrefixMatch_Sequential
测试顺序操作（先插入所有模式，再执行所有查询）。

### 6. BenchmarkPrefixMatch_Parallel_Search
测试并行查询性能。

### 7. BenchmarkPrefixMatch_LongPatterns
测试长模式匹配性能（路径很长的情况）。

### 8. BenchmarkPrefixMatch_ShortPatterns
测试短模式匹配性能（简短的键）。

### 9. BenchmarkPrefixMatch_CommonPrefix
测试具有公共前缀的模式（最坏情况）。

## 运行 Benchmark

```bash
# 运行所有 benchmark
go test -bench=. -benchmem -benchtime=3s prefixmatch_bench_test.go prefixmatch.go

# 运行特定 benchmark
go test -bench=BenchmarkPrefixMatch_Search_Mixed -benchmem

# 生成性能分析文件
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

# 查看性能分析
go tool pprof cpu.prof
```

## 典型性能数据

基于 Apple M1 Pro 的测试结果：

| 测试项                   | 操作数/秒    | 时间/操作 | 内存分配 |
| ------------------------ | ------------ | --------- | -------- |
| Insert (150个)           | ~50K ops     | ~20µs     | 31KB     |
| Search Hit               | ~20M ops     | ~62ns     | 0 B      |
| Search Miss              | ~85M ops     | ~14ns     | 0 B      |
| **Search Mixed (80/20)** | **~25M ops** | **~50ns** | **0 B**  |
| Parallel Search          | ~160M ops    | ~6ns      | 0 B      |
| Long Patterns            | ~20M ops     | ~60ns     | 0 B      |
| Short Patterns           | ~24M ops     | ~49ns     | 0 B      |

## 关键发现

1. **查询性能优异**: 混合查询场景下，单次查询仅需 50ns
2. **零内存分配**: 查询操作不产生任何内存分配
3. **未命中更快**: 未命中查询比命中查询快约 4 倍（14ns vs 62ns）
4. **并行扩展性好**: 并行查询性能提升显著
5. **模式长度影响小**: 长短模式的性能差异不大

## 与其他数据结构对比

相比于简单的 `map[string]int`:
- 前缀匹配能力：✅ (map: ❌)
- 内存效率：✅ 路径压缩节省空间
- 查询性能：~50ns vs ~10ns (略慢但可接受)
- 使用场景：路由匹配、URL 分发、权限路径匹配

## 验证测试

运行 `TestGenerateTestData` 验证测试数据的正确性：

```bash
go test -v -run TestGenerateTestData
```

该测试会验证：
- ✅ 生成 150 个唯一模式
- ✅ 120 个查询 (80%) 能命中
- ✅ 30 个查询 (20%) 不能命中
- ✅ 实际命中率 = 100%
- ✅ 实际未命中率 = 100%

oos: darwin
goarch: arm64
pkg: github.com/ethanvc/study/algo/prefixmatch
cpu: Apple M1 Pro
BenchmarkPrefixMatch_Insert-10                     57048             19208 ns/op      31312 B/op         396 allocs/op
BenchmarkPrefixMatch_Search_Hit-10              20710647                59.46 ns/op       0 B/op           0 allocs/op
BenchmarkPrefixMatch_Search_Miss-10             87189234                14.22 ns/op       0 B/op           0 allocs/op
BenchmarkPrefixMatch_Search_Mixed-10            24229064                47.57 ns/op       0 B/op           0 allocs/op
BenchmarkPrefixMatch_Sequential-10                 45226             26760 ns/op      31312 B/op         396 allocs/op
BenchmarkPrefixMatch_Parallel_Search-10         170432817                7.727 ns/op       0 B/op          0 allocs/op
BenchmarkPrefixMatch_LongPatterns-10            19256562                61.81 ns/op       0 B/op           0 allocs/op
BenchmarkPrefixMatch_ShortPatterns-10           24126786                48.96 ns/op       0 B/op           0 allocs/op
BenchmarkPrefixMatch_CommonPrefix-10            20115271                57.94 ns/op       0 B/op           0 allocs/op