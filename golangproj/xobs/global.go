package xobs

import (
	"crypto/rand"
	"encoding/hex"
)

var generateTraceIdFunc = GenerateTraceId
var generateSpanIdFunc = GenerateSpanId

func SetGenerateTraceIdFunc(f func() string) {
	generateTraceIdFunc = f
}

func SetGenerateSpanIdFunc(f func(rootSpan bool) string) {
	generateSpanIdFunc = f
}

func GenerateTraceId() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func GenerateSpanId(rootSpan bool) string {
	if rootSpan {
		return "0"
	}
	var buf [8]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}
