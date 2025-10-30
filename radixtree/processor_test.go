package radixtree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGinPatternProcessor_SplitPattern(t *testing.T) {
	var processor GinPatternProcessor
	nodes, err := processor.SplitPattern("/abc/:id")
	require.NoError(t, err)
	require.Len(t, nodes, 2)
}
