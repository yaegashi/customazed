package inpututil

import (
	"io/ioutil"
	"os"
	"strings"

	"muzzammil.xyz/jsonc"
)

func Read(input string) ([]byte, error) {
	if input == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(input)
}

func ReadJSONC(input string) ([]byte, error) {
	if input == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	b, err := ioutil.ReadFile(input)
	if err != nil && strings.HasSuffix(input, ".json") {
		b, err = ioutil.ReadFile(input + "c")
	}
	return b, err
}

func UnmarshalJSONC(input string, v interface{}) error {
	b, err := ReadJSONC(input)
	if err != nil {
		return err
	}
	return jsonc.Unmarshal(b, v)
}
