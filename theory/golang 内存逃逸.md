# slog.Info 内存逃逸分析
```golang
// 1
slog.Info("benchmark message",
				"url", "https://example.com/api",
				"status", 200,
				"err", errSample,
			)

// 2
slog.Info("benchmark message",
				"url", "https://example.com/api",
				"status", 200,
				"err", errSample,
				"err1", errSample,
				"err2", errSample,
			)
```

2会产生内存分配，1不会。原因如下：
1. 内联展开。
2. Record内部有5个item，用于保存日志项，record通过拷贝传值，不会进入堆。
3. 2 超过了5个，会导致分配一个slice保存多出来的kv数据。