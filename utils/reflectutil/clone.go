package reflectutil

import (
	"fmt"
	"reflect"
)

func Clone(x interface{}) interface{} {
	s := reflect.ValueOf(x)
	d := reflect.New(s.Type())
	_copy(s, d.Elem())
	return d.Elem().Interface()
}

func Copy(src interface{}, dst interface{}) {
	_copy(reflect.ValueOf(src), reflect.ValueOf(dst))
}

func _copy(src reflect.Value, dst reflect.Value) {
	if src.Type() != dst.Type() {
		panic(fmt.Errorf("type mismatch: src %s != dst %s", src.Type().String(), dst.Type().String()))
	}
	switch src.Kind() {
	default:
		dst.Set(src)
	case reflect.Ptr:
		if !src.IsNil() {
			tmp := reflect.New(src.Elem().Type())
			_copy(src.Elem(), tmp.Elem())
			dst.Set(tmp)
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			_copy(src.Field(i), dst.Field(i))
		}
	case reflect.Map:
		if !src.IsNil() {
			tmp := reflect.MakeMap(src.Type())
			itr := src.MapRange()
			for itr.Next() {
				val := reflect.New(itr.Value().Type()).Elem()
				_copy(itr.Value(), val)
				tmp.SetMapIndex(itr.Key(), val)
			}
			dst.Set(tmp)
		}
	case reflect.Slice:
		if !src.IsNil() {
			tmp := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
			for i := 0; i < src.Len(); i++ {
				_copy(src.Index(i), tmp.Index(i))
			}
			dst.Set(tmp)
		}
	}
}
