package logjson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogJson_Basic(t *testing.T) {
	type Abc struct {
		Name string `json:"name" logjson:"md5"`
	}
	lj := NewLogJson()
	buf, err := lj.Marshal(Abc{
		Name: "test",
	})
	require.NoError(t, err)
	require.Equal(t, `{"Name":""}`, string(buf))
}
