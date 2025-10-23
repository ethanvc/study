package envchange

import (
	"os"
	"testing"
)

func Test_GetEnv(t *testing.T) {
	os.Getenv("XXXXX")
}
