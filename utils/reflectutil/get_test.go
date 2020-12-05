package reflectutil_test

import (
	"testing"

	"github.com/yaegashi/customazed/utils/reflectutil"
)

type SubStruct struct {
	StructString string `json:"struct_string,omitempty"`
	StructInt    int    `json:"struct_int,omitempty"`
}

type MainStruct struct {
	Map    map[string]interface{} `json:"map,omitempty"`
	Struct SubStruct              `json:"struct,omitempty"`
	String string                 `json:"string,omitempty"`
	Int    int                    `json:"int,omitempty"`
}

func TestGet(t *testing.T) {
	mapData := map[string]interface{}{
		"MapString": "MapString",
		"MapInt":    123,
	}
	structData := MainStruct{
		Map: mapData,
		Struct: SubStruct{
			StructString: "StructString",
			StructInt:    456,
		},
		String: "String",
		Int:    789,
	}
	cases := []struct {
		val interface{}
		key string
		err bool
		exp string
	}{
		{
			val: structData,
			key: "String",
			exp: "String",
		},
		{
			val: structData,
			key: "Int",
			exp: "789",
		},
		{
			val: structData,
			key: "Struct.StructString",
			exp: "StructString",
		},
		{
			val: structData,
			key: "Struct.StructInt",
			exp: "456",
		},
		{
			val: structData,
			key: "struct.struct_int",
			exp: "456",
		},
		{
			val: structData,
			key: "struct.structint",
			err: true,
		},
		{
			val: structData,
			key: "Map.MapString",
			exp: "MapString",
		},
		{
			val: structData,
			key: "Map.MapInt",
			exp: "123",
		},
		{
			val: structData,
			key: "map.mapint",
			err: true,
		},
		{
			val: structData,
			key: "A",
			err: true,
		},
		{
			val: structData,
			key: "Struct.A",
			err: true,
		},
		{
			val: structData,
			key: "Map.A",
			err: true,
		},
		{
			val: structData,
			key: "Map",
			err: true,
		},
		{
			val: structData,
			key: "Map.MapString.A",
			err: true,
		},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			x, err := reflectutil.Get(c.val, c.key)
			if c.err {
				if err == nil {
					t.Errorf("got %q, want error", x)
				}
			} else {
				if err != nil {
					t.Error(err)
				}
				if x != c.exp {
					t.Errorf("got %q, want %q", x, c.exp)
				}
			}
		})
	}
}
