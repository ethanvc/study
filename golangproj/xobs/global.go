package xobs

import (
	"crypto/rand"
	"encoding/hex"
)

var GenerateTraceIdFunc = GenerateTraceId
var GenerateSpanIdFunc = GenerateSpanId

func GenerateTraceId() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func GenerateSpanId() string {
	var buf [8]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}
