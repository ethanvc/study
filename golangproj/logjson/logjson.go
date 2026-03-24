package logjson

import "github.com/ethanvc/study/golangproj/logjson/internal/gostd/encoding/json/v2"

type LogJson struct{}

func NewLogJson() *LogJson {
	return &LogJson{}
}

func (lj *LogJson) AddMarshaler(key string, f MarshalFunc) {
}

func (lj *LogJson) Marshal(v any, opts ...Options) ([]byte, error) {
	return json.Marshal(v, opts...)
}

type MarshalFunc func(encoder *Encoder, v any) error
