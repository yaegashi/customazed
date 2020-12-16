package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/yaegashi/customazed/utils/reflectutil"
)

type TemplateVariable struct {
	err     error
	cache   map[string]string
	ref     map[string]bool
	funcMap template.FuncMap
}

func (app *App) NewTemplateVariable(su StorageUploader) *TemplateVariable {
	if su == nil {
		su = DisabledStorageUploader("upload: no configuration")
	}
	tv := &TemplateVariable{
		cache: map[string]string{},
		ref:   map[string]bool{},
	}
	tv.funcMap = template.FuncMap{
		"upload": tv.NewFunc("upload", func(key string) (string, error) {
			return su.Add(key)
		}),
		"cfg": tv.NewFunc("cfg", func(key string) (string, error) {
			return reflectutil.Get(app.ConfigLoad, key)
		}),
		"var": tv.NewFunc("var", func(key string) (string, error) {
			return reflectutil.Get(app.ConfigLoad.Variables, key)
		}),
		"env": tv.NewFunc("env", func(key string) (string, error) {
			if val, ok := os.LookupEnv(key); ok {
				return val, nil
			}
			return "", fmt.Errorf("Environment variable %q not found", key)
		}),
		"hash": func(s ...string) string { return app.HashID(s...) },
	}
	tv.funcMap["id"] = func() string { return tv.funcMap["cfg"].(func(string) string)("id") }
	tv.funcMap["prefix"] = func() string { return tv.funcMap["cfg"].(func(string) string)("storage.prefix") }
	return tv
}

func (tv *TemplateVariable) NewFunc(fName string, fCall func(string) (string, error)) func(string) string {
	return func(key string) string {
		cacheKey := fName + ":" + key
		if val, ok := tv.cache[cacheKey]; ok {
			return val
		}
		var (
			str string
			err error
		)
		if tv.ref[cacheKey] {
			err = fmt.Errorf("Cyclic reference %q", cacheKey)
		} else {
			str, err = fCall(key)
		}
		if err != nil {
			if tv.err == nil {
				tv.err = err
			}
			return fmt.Sprintf("<ERROR:%s>", err)
		}
		tv.ref[cacheKey] = true
		str, err = tv.Execute(str)
		if err != nil {
			if tv.err == nil {
				tv.err = err
			}
			return fmt.Sprintf("<ERROR:%s>", err)
		}
		tv.ref[cacheKey] = false
		tv.cache[cacheKey] = str
		return str
	}
}

func (tv *TemplateVariable) Execute(in string) (string, error) {
	tmpl, err := template.New("template").Funcs(tv.funcMap).Parse(in)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, tv)
	if err != nil {
		return "", err
	}
	if tv.err != nil {
		return "", tv.err
	}
	out := buf.String()
	return out, nil
}

func (tv *TemplateVariable) Resolve(v interface{}) error {
	var tmpErr error
	reflectutil.WalkSet(v, func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			t, err := tv.Execute(s)
			if tmpErr == nil && err != nil {
				tmpErr = err
				return ""
			}
			return t
		}
		return v
	})
	return tmpErr
}
