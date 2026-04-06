package json

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/ethanvc/study/golangproj/logjson/internal/gostd/encoding/json/internal/jsonopts"
	"github.com/ethanvc/study/golangproj/logjson/internal/gostd/encoding/json/jsontext"
)

func logjsonWrapArshaler(f *structField, t reflect.Type) {
	if !f.md5 {
		return
	}
	origFncs := f.fncs
	wrapped := &arshaler{
		marshal: func(enc *jsontext.Encoder, va addressableValue, mo *jsonopts.Struct) error {
			var data []byte
			switch va.Kind() {
			case reflect.String:
				data = []byte(va.String())
			case reflect.Slice:
				data = va.Bytes()
			default:
				return origFncs.marshal(enc, va, mo)
			}
			h := md5.Sum(data)
			return enc.WriteToken(jsontext.String(fmt.Sprintf("len=%d,%s", len(data), hex.EncodeToString(h[:]))))
		},
		unmarshal:  origFncs.unmarshal,
		nonDefault: true,
	}
	f.fncs = wrapped
}
