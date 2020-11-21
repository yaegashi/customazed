package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
)

type TemplateVariable struct {
	ctx context.Context
	app *App
	err error
	*AppConfig
}

func (app *App) TemplateExecute(ctx context.Context, in []byte) ([]byte, error) {
	tmplVars := TemplateVariable{
		ctx:       ctx,
		app:       app,
		AppConfig: &app.Config,
	}
	tmplFuncMap := template.FuncMap{
		"upload": tmplVars.Upload,
	}
	tmpl, err := template.New("template").Funcs(tmplFuncMap).Parse(string(in))
	if err != nil {
		return nil, err
	}
	outBuf := &bytes.Buffer{}
	err = tmpl.Execute(outBuf, tmplVars)
	if err != nil {
		return nil, err
	}
	return outBuf.Bytes(), nil
}

func (tv *TemplateVariable) Upload(path string) string {
	if tv.err != nil {
		return "<CANCELED>"
	}
	ctx, app, err := tv.ctx, tv.app, tv.err
	app.Logf("Uploading %s", path)
	u, err := app.StorageUpload(ctx, path)
	if err != nil {
		tv.err = err
		return fmt.Sprintf("<ERROR:%s>", err)
	}
	return u
}
