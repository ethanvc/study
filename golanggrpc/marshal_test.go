package golanggrpc

import (
	"github.com/ethanvc/study/golangproj/logjson/internal/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unmarshal(t *testing.T) {
	{
		var v StructWithData
		buf := `{"data":3}`
		bufBytes := []byte(buf)
		err := json.Unmarshal(bufBytes, &v)
		require.NoError(t, err)
	}
	{
		var v StructWithData
		buf := `{"data":"3"}`
		bufBytes := []byte(buf)
		err := json.Unmarshal(bufBytes, &v)
		require.NoError(t, err)
	}

}
