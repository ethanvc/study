package logjson

import "github.com/ethanvc/study/golangproj/logjson/internal/json/v2"

type LogJson struct{}

func NewLogJson() *LogJson {
	return &LogJson{}
}

func (lj *LogJson) Marshal(v any, opts ...Options) ([]byte, error) {
	return json.Marshal(v, opts...)
}

func (lj *LogJson) MarshalAsStr(v any, opts ...Options) (string, error) {
	buf, err := lj.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
