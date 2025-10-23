package golangjson

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func Test_BatchGet(t *testing.T) {
	jsonStr := `{
		"results": [
			{
				"zikey": "value1",
				"name": "item1"
			},
			{
				"zikey": "value2",
				"name": "item2"
			},
			{
				"zikey": "value3",
				"name": "item3"
			}
		]
	}`

	// 使用 gjson 的路径表达式获取所有 "zikey" 的值
	// "results.#.zikey" 表示：进入 "results" 数组，遍历所有元素（#），然后获取每个元素的 "zikey" 字段
	result := gjson.Get(jsonStr, "results.#.zikey")

	// result.Array() 返回一个 gjson.Result 类型的切片
	require.EqualValues(t, 3, len(result.Array()))
}
