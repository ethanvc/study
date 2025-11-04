package golanglimit

type GoSchedBreaker struct{}

func (b GoSchedBreaker) Break() bool {}
