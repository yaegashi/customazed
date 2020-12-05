package reflectutil

import (
	"fmt"
	"reflect"
	"strings"
)

func recursiveGet(v reflect.Value, kSlice []string) (interface{}, error) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if len(kSlice) == 0 {
		switch v.Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice:
			return "", fmt.Errorf("Unable to get value of %#v", v.Interface())
		default:
			return v.Interface(), nil
		}
	}
	k := kSlice[0]
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			g, _ := f.Tag.Lookup("json")
			g = strings.Split(g, ",")[0]
			if f.Name == k || g == k {
				return recursiveGet(v.Field(i), kSlice[1:])
			}
		}
		return "", fmt.Errorf("Key %q not found in %#v", k, v.Interface())
	case reflect.Map:
		mapV := v.MapIndex(reflect.ValueOf(k))
		if mapV.IsValid() {
			return recursiveGet(mapV, kSlice[1:])
		}
		return "", fmt.Errorf("Key %q not found in %#v", k, v.Interface())
	default:
		return "", fmt.Errorf("Unable to get key %q value of %#v", k, v.Interface())
	}
}

func splitKey(key string) []string {
	return strings.Split(key, ".")
}

func Get(val interface{}, key string) (string, error) {
	i, err := recursiveGet(reflect.ValueOf(val), splitKey(key))
	if err != nil {
		return "", err
	}
	return fmt.Sprint(i), nil
}
