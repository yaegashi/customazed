package reflectutil_test

import (
	"encoding/json"
	"testing"

	"github.com/yaegashi/customazed/utils/reflectutil"
)

type walkA struct {
	I   int
	S   string
	PI  *int
	PS  *string
	SI  []int
	SS  []string
	SPI []*int
	SPS []*string
	MIS map[int]string
}

type walkB struct {
	A  walkA
	PA *walkA
	I  interface{}
	PI interface{}
}

func TestWalk(t *testing.T) {
	i := 42
	pi := &i
	s := "abc"
	ps := &s
	a := walkA{
		I:   i,
		S:   s,
		PI:  pi,
		PS:  ps,
		SI:  []int{i, i, i},
		SS:  []string{s, s, s},
		SPI: []*int{pi, pi, pi},
		SPS: []*string{ps, ps, ps},
		MIS: map[int]string{1: "a", 2: "b", 3: "c"},
	}
	b := walkB{
		A:  a,
		PA: &a,
		I:  a,
		PI: &a,
	}
	reflectutil.WalkSet(b, func(v interface{}) interface{} {
		if i, ok := v.(int); ok {
			return i + 1
		}
		if s, ok := v.(string); ok {
			return s + "+"
		}
		return v
	})
	jb, _ := json.Marshal(b)
	t.Log(string(jb))
}
