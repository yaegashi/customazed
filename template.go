package main

import (
	"bytes"
	"context"
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

func (app *App) TemplateExecute(ctx context.Context, in []byte) ([]byte, error) {
	if app._TemplateCache == nil {
		app._TemplateCache = map[string]string{}
	}
	if app._TemplateRef == nil {
		app._TemplateRef = map[string]bool{}
	}
	tv := &TemplateVariable{
		cache: app._TemplateCache,
		ref:   app._TemplateRef,
	}
	tv.funcMap = template.FuncMap{
		"upload": tv.NewFunc("upload", func(key string) (string, error) {
			app.Logf("Uploading %s", key)
			return app.StorageUpload(ctx, key)
		}),
		"cfg": tv.NewFunc("cfg", func(key string) (string, error) {
			return reflectutil.Get(app.Config, key)
		}),
		"var": tv.NewFunc("var", func(key string) (string, error) {
			return reflectutil.Get(app.Config.Variables, key)
		}),
		"env": tv.NewFunc("env", func(key string) (string, error) {
			if val, ok := os.LookupEnv(key); ok {
				return val, nil
			}
			return "", fmt.Errorf("Environment variable %q not found", key)
		}),
	}
	out, err := tv.Execute(string(in))
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
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

func (app *App) TemplateResolve(ctx context.Context, val interface{}) (interface{}, error) {
	var tmpErr error
	v := reflectutil.Clone(val)
	reflectutil.WalkSet(v, func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			b, err := app.TemplateExecute(ctx, []byte(s))
			if tmpErr == nil && err != nil {
				tmpErr = err
				return ""
			}
			return string(b)
		}
		return v
	})
	if tmpErr != nil {
		return nil, tmpErr
	}
	return v, nil
}
