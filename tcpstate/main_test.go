package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AddressReslove(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "*:53")
	require.NoError(t, err)
	_ = addr
}
