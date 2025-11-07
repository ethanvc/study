package main

import "runtime"

func main() {
	mainGosched()
}

//go:noinline
func mainGosched() {
	runtime.GOMAXPROCS(2)
	for {
		runtime.Gosched()
	}
}
