package json

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
	"strconv"

	"github.com/ethanvc/study/golangproj/logjson/internal/gostd/encoding/json/internal/jsonopts"
	"github.com/ethanvc/study/golangproj/logjson/internal/gostd/encoding/json/jsontext"
)

func logjsonWrapArshaler(f *structField) {
	if !f.md5 {
		return
	}
	origFncs := f.fncs
	wrapped := &arshaler{
		marshal: func(enc *jsontext.Encoder, va addressableValue, mo *jsonopts.Struct) error {
			// "len=" (4) + max int64 digits (20) + "," (1) + md5 hex (32) = 57
			var buf [57]byte
			b := append(buf[:0], "len="...)

			var h [md5.Size]byte
			switch va.Kind() {
			case reflect.String:
				s := va.String()
				b = strconv.AppendInt(b, int64(len(s)), 10)
				h = md5.Sum([]byte(s))
			case reflect.Slice:
				data := va.Bytes()
				b = strconv.AppendInt(b, int64(len(data)), 10)
				h = md5.Sum(data)
			default:
				return origFncs.marshal(enc, va, mo)
			}
			b = append(b, ',')
			b = hex.AppendEncode(b, h[:])
			return enc.WriteToken(jsontext.String(string(b)))
		},
		unmarshal:  origFncs.unmarshal,
		nonDefault: true,
	}
	f.fncs = wrapped
}
