package golanggrpc

import (
	"testing"
	"time"
)

func test_StartTestServer(t *testing.T) {
	_, _, cancel := createPair(t)
	defer cancel()
	time.Sleep(5 * time.Minute)
}
