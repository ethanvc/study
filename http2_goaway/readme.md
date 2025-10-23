以下分析的前置条件：
1. http2协议。

通过配置 `keepalive_requests 2;`，nginx在处理完第二个请求后，会发送goaway帧，如果服务端并发度比较高，可能会刚好使用即将关闭的连接。
此时报错信息为：
```
Post "https://localhost": http2: Transport: cannot retry err [http2: Transport received Server's graceful shutdown 
GOAWAY] after Request.Body was written; define Request.GetBody to avoid this error
```

在http2协议中，由于是并发复用连接，nginx无法使用返回头 Connection:close 告知client，连接即将被关闭。