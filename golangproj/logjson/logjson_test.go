package logjson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogJson_Basic(t *testing.T) {
	type Abc struct {
		Name   string            `json:"name" logjson:"md5"`
		Values map[string]string `json:"values" logjson:"map,password:md5,sign_key:md5"`
	}
	lj := NewLogJson()
	buf, err := lj.MarshalAsStr(Abc{
		Name: "test",
	})
	require.NoError(t, err)
	require.Equal(t, `{"Name":"md5,4:098f6bcd4621d373cade4e832627b4f6"}`, buf)
}
