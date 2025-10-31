package radixtree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRadixTree_Insert(t *testing.T) {
	var tree RadixTree[PlainProcessor, int]
	tree.MustInsert("/abc", 3)
	tree.MustInsert("/abc/def", 4)
}

func Test_DuplicateInsert(t1 *testing.T) {
	var tree RadixTree[PlainProcessor, int]
	err := tree.Insert("/abc", 3)
	require.NoError(t1, err)
	err = tree.Insert("/abc", 4)
	require.EqualError(t1, err, "pattern already exist: /abc")
}

func TestRadixTree_1(t *testing.T) {
	var tree RadixTree[GinPatternProcessor, int]
	tree.MustInsert("/abc/:id", 3)
	tree.MustInsert("/abc/bcd", 4)
	n, params, err := tree.Search("/abc/bcde", nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(params))
	require.Equal(t, "/abc/:id", n.pattern)

	n, params, err = tree.Search("/abc/bcd", nil)
	require.NoError(t, err)
	require.EqualValues(t, 0, len(params))
	require.Equal(t, "/abc/bcd", n.pattern)

}

func Test_PatternParam(t *testing.T) {
	tree := NewGinRadixTree[int]()
	tree.MustInsert("/api3/bcde/3", 5)
	tree.MustInsert("/api3/*rest", 6)
	n, params, err := tree.Search("/api3/bcde/3", nil)
	require.NoError(t, err)
	require.EqualValues(t, 0, len(params))
	require.Equal(t, 5, n.GetValue())

	n, params, err = tree.Search("/api3/bcde/33", nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(params))
	require.Equal(t, 6, n.GetValue())
}
