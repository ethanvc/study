# 更新标准库的json包
1. 拷贝标准库的json包到internal目录下。
2. 字符串替换：
```
"encoding/json->"github.com/ethanvc/study/golangproj/logjson/internal/json

//go: build goexperiment.jsonv2-><blank>
```

# tag 串设计
完整语法：
key1:val1;key2:val2;val3;

可以看到，支持kv结构和value列表。