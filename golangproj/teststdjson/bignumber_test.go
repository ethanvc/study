package teststdjson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type Order struct {
	Amount int64 `json:"amount"`
}

func Test_bigNumber(t *testing.T) {
	s := `{"amount": 9007199254740993}`
	var order Order
	err := json.Unmarshal([]byte(s), &order)
	require.NoError(t, err)
	require.Equal(t, int64(9007199254740993), order.Amount)
}
