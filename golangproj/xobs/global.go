package xobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"
)

var generateTraceIdFunc = GenerateTraceId
var generateSpanIdFunc = GenerateSpanId

var defaultSpan = newDefaultSpan()

var defaultHandler = NewJsonHandler(os.Stdout)

// for test only
var sNow = time.Now

func SetDefaultSpan(span *Span) {
	defaultSpan = span
}

func SetGenerateTraceIdFunc(f func() string) {
	generateTraceIdFunc = f
	defaultSpan = newDefaultSpan()
}

func SetGenerateSpanIdFunc(f func(rootSpan bool) string) {
	generateSpanIdFunc = f
	defaultSpan = newDefaultSpan()
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

func newDefaultSpan() *Span {
	return NewSpan(context.Background(), &SpanConfig{
		Name: "default",
	})
}
