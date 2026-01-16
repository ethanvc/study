package teststdjson

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Format(t *testing.T) {
	type Abc struct {
		Val []byte `json:"val,format:base64"`
	}
	val := Abc{[]byte("abc")}
	buf, err := json.Marshal(val)
	require.NoError(t, err)
	require.Equal(t, `{"val":"YWJj"}`, string(buf))
}
