package filechange

import (
	"os"
	"testing"
)

func Test_FileChange(t *testing.T) {
	filename := "xx.txt"
	// 要写入的内容
	data := []byte("hello")

	// os.WriteFile 会创建一个文件（如果不存在），截断文件（如果存在），写入数据，然后关闭文件。
	// 0644 是文件权限 (rw-r--r--)
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatal(err)
	}
}
