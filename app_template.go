package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppTemplate struct {
	*App
	Input  string
	Output string
}

func (app *App) AppTemplateCmder() cmder.Cmder {
	return &AppTemplate{App: app}
}

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

func (app *AppTemplate) RunE(cmd *cobra.Command, args []string) error {
	var err error
	var inBuf []byte
	if app.Input == "-" {
		inBuf, err = ioutil.ReadAll(os.Stdin)
	} else {
		inBuf, err = ioutil.ReadFile(app.Input)
	}
	if err != nil {
		return err
	}

	_, err = app.ARMToken()
	if err != nil {
		return err
	}
	_, err = app.StorageToken()
	if err != nil {
		return err
	}

	ctx := context.Background()
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

	outBuf, err := app.TemplateExecute(ctx, inBuf)
	if err != nil {
		return err
	}

	if app.Output == "-" {
		_, err = os.Stdout.Write(outBuf)
	} else {
		err = ioutil.WriteFile(app.Output, outBuf, 0644)
	}
	if err != nil {
		return err
	}

	return nil
}
