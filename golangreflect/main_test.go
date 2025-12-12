package golangreflect

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func Test_SetPrivateField(t *testing.T) {
	type Abc struct {
		a int
	}
	instance := Abc{a: 1}

	// 1. 获取指向 instance 的指针的 reflect.Value，并解引用 (.Elem())
	v := reflect.ValueOf(&instance).Elem()

	// 2. 找到私有字段 'a'
	fieldA := v.FieldByName("a")

	// 3. 使用 unsafe 获取字段的内存地址
	ptr := unsafe.Pointer(fieldA.UnsafeAddr())

	// 4. 创建一个新的、可设置的 Value，指向该内存地址
	settableField := reflect.NewAt(fieldA.Type(), ptr).Elem()

	// 5. 设置新值 3
	if settableField.Kind() == reflect.Int {
		settableField.SetInt(3)
	}
	require.Equal(t, 3, instance.a)
}

func Test_SetPublicField(t *testing.T) {
	type Abc struct {
		A int
	}
	instance := Abc{A: 1}

	// 1. 获取指向 instance 的指针的 reflect.Value，并解引用 (.Elem())
	// 只有通过指针获取的 Value 才是可设置的 (Set-able)
	v := reflect.ValueOf(&instance).Elem()

	// 2. 找到可导出字段 'A'
	fieldA := v.FieldByName("A")

	// 3. 检查字段是否可设置 (CanSet)
	// 对于可导出的字段，CanSet 会返回 true
	if fieldA.CanSet() {
		// 4. 设置新值 3
		if fieldA.Kind() == reflect.Int {
			fieldA.SetInt(3)
		}
	}
	require.Equal(t, 3, instance.A)
}
