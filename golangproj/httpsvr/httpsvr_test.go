package httpsvr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer_Basic(t *testing.T) {
}

func Test_validateAndParseFunc(t *testing.T) {
	f := func(context.Context, *any) (*any, error) {
		return nil, nil
	}
	reqType, respType, err := validateAndParseFunc(f)
	require.NoError(t, err)
	require.Equal(t, "interface {}", reqType.String())
	require.Equal(t, "interface {}", respType.String())
}
