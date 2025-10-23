package golangjson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_jsonMerge(t *testing.T) {
	type Abc struct {
		Val  int
		Msg  string
		List []string
	}
	s := `{"Val":3, "Msg":"xxx", "List":["Hello"]}`
	var abc Abc
	err := json.Unmarshal([]byte(s), &abc)
	require.NoError(t, err)
	s = `{"List":["Hello2"]}`
	err = json.Unmarshal([]byte(s), &abc)
	require.NoError(t, err)
	require.Equal(t, 1, len(abc.List))
	require.Equal(t, "Hello2", abc.List[0])
	require.Equal(t, 3, abc.Val)
}
