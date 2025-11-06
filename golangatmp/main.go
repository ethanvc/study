package main

import "runtime"

func main() {
	mainGosched()
}

//go:noinline
func mainGosched() {
	for {
		runtime.Gosched()
	}
}
