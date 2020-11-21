// +build !windows

package store

import (
	"fmt"
	"os"
	"testing"
)

func TestLocation(t *testing.T) {
	tests := []struct {
		r          bool
		h, d, i, o string
		u          bool
	}{
		{r: true, h: "/home/user", d: "~/.config", i: "token.json", o: "/home/user/.config/token.json", u: false},
		{r: true, h: "/home/user", d: "~/.config", i: "/token.json", o: "/token.json", u: false},
		{r: true, h: "/home/user", d: "~/.config", i: "./token.json", o: "./token.json", u: false},
		{
			r: true,
			h: "/home/user",
			d: "~/.config",
			i: "https://storage.blob.core.windows.net/container/token.json?key=secret",
			o: "https://storage.blob.core.windows.net/container/token.json?...",
			u: true,
		},
		{
			r: false,
			h: "/home/user",
			d: "~/.config",
			i: "https://storage.blob.core.windows.net/container/token.json?key=secret",
			o: "https://storage.blob.core.windows.net/container/token.json?key=secret",
			u: true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			os.Setenv("HOME", tt.h)
			s, err := NewStore(tt.d)
			if err != nil {
				t.Fatal(err)
			}
			o, u := s.Location(tt.i, tt.r)
			if tt.o != o {
				t.Errorf("Path mismatch want %q got %q", tt.o, o)
			}
			if tt.u != u {
				t.Errorf("URL mismatch want %v got %v", tt.u, u)
			}
		})
	}
}
