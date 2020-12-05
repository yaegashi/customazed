package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppConfig struct {
	*App
}

func (app *App) AppConfigCmder() cmder.Cmder {
	return &AppConfig{App: app}
}

func (app *AppConfig) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "config",
		Short:        "Configuration",
		SilenceUsage: true,
	}
	return cmd
}
