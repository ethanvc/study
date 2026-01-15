package teststdjson

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Test_ProtoJson(t *testing.T) {
	protoVal := wrapperspb.String(`"hello world"`)
	var value struct {
		// GoStruct does not implement proto.Message and
		// should use the default behavior of the "json" package.
		GoStruct struct {
			Name string
			Age  int
		}

		// ProtoMessage implements proto.Message and
		// should be handled using protojson.Marshal.
		ProtoMessage *wrapperspb.StringValue
	}
	value.ProtoMessage = protoVal

	// Marshal using protojson.Marshal for proto.Message types.
	b, err := json.Marshal(&value,
		// Use protojson.Marshal as a type-specific marshaler.
		json.WithMarshalers(json.MarshalFunc(protojson.Marshal)))
	require.NoError(t, err)

	// Unmarshal using protojson.Unmarshal for proto.Message types.
	err = json.Unmarshal(b, &value,
		// Use protojson.Unmarshal as a type-specific unmarshaler.
		json.WithUnmarshalers(json.UnmarshalFunc(protojson.Unmarshal)))
	require.NoError(t, err)
}
