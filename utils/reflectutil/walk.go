package reflectutil

import (
	"reflect"
)

// WalkSet walks data values and call function to modify each of them
func WalkSet(val interface{}, fn func(interface{}) interface{}) {
	_walkSet(reflect.ValueOf(val), fn)
}

func _walkSet(v reflect.Value, fn func(interface{}) interface{}) {
	switch v.Kind() {
	default:
		i := fn(v.Interface())
		if v.CanSet() {
			v.Set(reflect.ValueOf(i))
		}
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			_walkSet(v.Elem(), fn)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			_walkSet(v.Field(i), fn)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			_walkSet(v.Index(i), fn)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			v.SetMapIndex(k, reflect.ValueOf(fn(v.MapIndex(k).Interface())))
		}
	}
}
