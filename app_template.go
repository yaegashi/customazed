package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppTemplate is app template command
type AppTemplate struct {
	*App
	Input  string
	Output string
}

// AppTemplateCmder returns Cmder for app template
func (app *App) AppTemplateCmder() cmder.Cmder {
	return &AppTemplate{App: app}
}

// Cmd returns Command for app template
func (app *AppTemplate) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "template",
		Aliases:      []string{"t"},
		Short:        "Customazed template",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&app.Input, "input", "i", "-", "input file path")
	cmd.Flags().StringVarP(&app.Output, "output", "o", "-", "output file path")
	return cmd
}

// RunE is main routine for app template
func (app *AppTemplate) RunE(cmd *cobra.Command, args []string) error {
	var err error
	var buf []byte
	if app.Input == "-" {
		buf, err = ioutil.ReadAll(os.Stdin)
	} else {
		buf, err = ioutil.ReadFile(app.Input)
	}
	if err != nil {
		return err
	}
	in := string(buf)

	ctx := context.Background()

	if !app.NoLogin {
		_, err = app.ARMToken()
		if err != nil {
			return err
		}
		_, err = app.StorageToken()
		if err != nil {
			return err
		}
		_, err = app.StorageAccount(ctx)
		if err != nil {
			return err
		}
		_, err = app.Identity(ctx)
		if err != nil {
			return err
		}
		_, err = app.Image(ctx)
		if err != nil {
			return err
		}
	}

	su := app.NewStorageUploader(ctx)
	tv := app.NewTemplateVariable(su)
	out, err := tv.Execute(in)
	if err != nil {
		return err
	}

	if app.Output == "-" {
		_, err = os.Stdout.WriteString(out)
	} else {
		app.Logf("Wrinting to %s", app.Output)
		err = ioutil.WriteFile(app.Output, []byte(out), 0644)
	}
	if err != nil {
		return err
	}

	app.Prompt("Files to upload: %d", su.Files())

	if su.Valid() && su.Files() > 0 {
		err = su.Execute(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
