package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppLogin is app login command
type AppLogin struct {
	*App
}

// AppLoginCmder returns Cmder for app login
func (app *App) AppLoginCmder() cmder.Cmder {
	return &AppLogin{App: app}
}

// Cmd returns Command for app login
func (app *AppLogin) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "login",
		Short:        "Force dev auth login",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app login
func (app *AppLogin) RunE(cmd *cobra.Command, args []string) error {
	_, err := app.AuthorizeDeviceFlow()
	return err
}
