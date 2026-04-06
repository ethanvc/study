package json

import (
	"fmt"
	"reflect"
	"strings"
)

type logjsonFieldOptions struct {
	md5 bool
}

func parseLogjsonFieldOptions(sf reflect.StructField, out *fieldOptions) error {
	tag, hasTag := sf.Tag.Lookup("logjson")
	if !hasTag {
		return nil
	}
	for _, opt := range strings.Split(tag, ",") {
		opt = strings.TrimSpace(opt)
		switch opt {
		case "md5":
			k := sf.Type.Kind()
			if k == reflect.String {
				out.md5 = true
			} else if k == reflect.Slice && sf.Type.Elem().Kind() == reflect.Uint8 {
				out.md5 = true
			} else {
				return fmt.Errorf("Go struct field %s has `logjson:\"md5\"` tag but type %s is not string or []byte", sf.Name, sf.Type)
			}
		case "":
		default:
			return fmt.Errorf("Go struct field %s has unknown `logjson` tag option %q", sf.Name, opt)
		}
	}
	return nil
}
