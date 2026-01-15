package golanggrpc

import (
	"encoding/json/v2"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func Test_ProtoJson(t *testing.T) {
	protoVal := &ProtoJsonMsg{
		Int64Val: math.MaxInt64,
	}
	type Value struct {
		Name string
		Age  int64
		// should be handled using protojson.Marshal.
		ProtoMessage *ProtoJsonMsg
	}
	var value Value
	value.Name = "John Doe"
	value.ProtoMessage = protoVal

	// Marshal using protojson.Marshal for proto.Message types.
	b, err := json.Marshal(&value,
		// Use protojson.Marshal as a type-specific marshaler.
		json.WithMarshalers(json.MarshalFunc(protojson.Marshal)))
	require.NoError(t, err)
	require.Equal(t, `{"Name":"John Doe","Age":0,"ProtoMessage":{"int64Val":"9223372036854775807"}}`, string(b))

	var newValue Value
	err = json.Unmarshal(b, &newValue,
		// Use protojson.Unmarshal as a type-specific unmarshaler.
		json.WithUnmarshalers(json.UnmarshalFunc(protojson.Unmarshal)))
	require.NoError(t, err)
	require.Equal(t, int64(math.MaxInt64), newValue.ProtoMessage.Int64Val)
}
