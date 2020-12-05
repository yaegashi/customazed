package reflectutil_test

import (
	"testing"

	"github.com/yaegashi/customazed/utils/reflectutil"
)

type cloneA struct {
	I  int
	S  string
	PI *int
	PS *string
}

type cloneB struct {
	A  cloneA
	PA *cloneA
}

type cloneC struct {
	SI  []int
	MSI map[string]int
}

func TestClone(t *testing.T) {
	i := 42
	pi := &i
	ppi := &pi
	t.Logf("%#v", reflectutil.Clone(i))
	t.Logf("%#v", reflectutil.Clone(pi))
	t.Logf("%#v", reflectutil.Clone(ppi))
	s := "abc"
	ps := &s
	pps := &ps
	t.Logf("%#v", reflectutil.Clone(s))
	t.Logf("%#v", reflectutil.Clone(ps))
	t.Logf("%#v", reflectutil.Clone(pps))
	a := cloneA{I: i, S: s, PI: pi, PS: ps}
	pa := &a
	t.Logf("%#v", reflectutil.Clone(a))
	t.Logf("%#v", reflectutil.Clone(pa))
	b := cloneB{A: a, PA: &a}
	pb := &b
	t.Logf("%#v", reflectutil.Clone(b))
	t.Logf("%#v", reflectutil.Clone(pb))
	t.Logf("%#v", reflectutil.Clone(cloneB{}))
	t.Logf("%#v", reflectutil.Clone(&cloneB{}))
	c := cloneC{SI: []int{1, 2, 3}, MSI: map[string]int{"a": 1, "b": 2, "c": 3}}
	t.Logf("%#v", reflectutil.Clone(c))
}
