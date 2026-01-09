# 更新标准库的json包
1. 拷贝标准库的json包到internal目录下。
2. 字符串替换：
```
"encoding/json->"github.com/ethanvc/study/golangproj/logjson/internal/json

//go: build goexperiment.jsonv2-><blank>
```