package golangtestcache_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ethanvc/study/golangtestcache"
)

/*
go test 和 go test . 是不同的，前者始终不会走cache
详情见 go help test（local directory mode/package list mode）

GODEBUG=gocachehash=1,gocachetest=1 go test -x . 2>&1 | tee test.log
GODEBUG=gocachetest=1 go test -x . 2>&1 | tee test.log

源代码的改动不一定会导致cache失效，应该是比较的构建物的hash。
TestMain不调用m.Run，就会无法cache。
*/

func TestMain(m *testing.M) {
	if _, ok := os.LookupEnv("xxx"); !ok {
		m.Run()
		return
	}
	m.Run()
}

func Test_Abc(t *testing.T) {
	golangtestcache.Func()
	fmt.Print(3, 4)
	fmt.Fprint(os.Stderr, 3, 4)
}
