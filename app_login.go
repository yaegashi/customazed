package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppLogin struct {
	*App
}

func (app *App) AppLoginCmder() cmder.Cmder {
	return &AppLogin{App: app}
}

func (app *AppLogin) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "login",
		Short:        "Force dev auth login",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func (app *AppLogin) RunE(cmd *cobra.Command, args []string) error {
	_, err := app.AuthorizeDeviceFlow()
	return err
}
