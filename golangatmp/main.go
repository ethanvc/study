package main

import "runtime"

func main() {
	mainGosched()
}

func mainGosched() {
	for {
		runtime.Gosched()
	}
}
