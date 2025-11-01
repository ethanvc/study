# radixtree
实现和gin类似的功能，gin采用数据搜索子节点，但是当子节点太多的时候，性能退化验证。
在实际生成环境，存在 indices="vasipbgmcjfouredqtVwhlRn" 的情况。但是从整体上来看，很少有这种场景，因此gin的选择仍然是最优的。