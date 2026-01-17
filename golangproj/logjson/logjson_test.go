package logjson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogJson_Basic(t *testing.T) {
	type Abc struct {
		// ignore output to log
		Password2 string `json:"password2" logjson:"ignore"`
		// secret must mask
		Password string `json:"password" logjson:"md5"`
		// some keys are secret
		Values map[string]string `json:"values" logjson:"map,password:md5,sign_key:md5"`
		// too long for log, only need check if right value
		FaceFeature []byte `json:"face_feature" logjson:"md5"`
	}
	lj := NewLogJson()
	lj.AddMarshaler("face_feature", func(encoder *Encoder, v any) error {
		return encoder.WriteToken(String("xxx"))
	})
	buf, err := marshal(lj, Abc{
		Password: "test",
		Values: map[string]string{
			"password": "test",
			"sign_key": "test",
			"other":    "xx",
		},
	})
	require.NoError(t, err)
	require.Equal(t, `{
    "name": "test",
    "values": {
        "password": "test",
        "sign_key": "test",
        "other": "xx"
    },
    "face_feature": ""
}`, buf)
}

func Test_ProtoStruct(t *testing.T) {
	val := &ProtoLogJsonTest{
		UserName: "test",
	}
	lj := NewLogJson()
	buf, err := marshal(lj, val)
	require.NoError(t, err)
	require.Equal(t, `{}`, buf)
}

func marshal(lj *LogJson, v any, opts ...Options) (string, error) {
	opts = append(opts, WithIndent("    "), Multiline(true))
	buf, err := lj.Marshal(v, opts...)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
