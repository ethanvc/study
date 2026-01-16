package logjson

import (
	"github.com/ethanvc/study/golangproj/logjson/internal/json"
	"github.com/ethanvc/study/golangproj/logjson/internal/json/jsontext"
)

type Options = json.Options

type Encoder = jsontext.Encoder

type Value = jsontext.Value
type Token = json.Token

var String = jsontext.String

var Multiline = jsontext.Multiline
var WithIndent = jsontext.WithIndent
