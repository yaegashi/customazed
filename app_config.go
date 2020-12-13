package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppConfig is app config command
type AppConfig struct {
	*App
}

// AppConfigCmder returns Cmder for app config
func (app *App) AppConfigCmder() cmder.Cmder {
	return &AppConfig{App: app}
}

// Cmd returns Command for app config
func (app *AppConfig) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "config",
		Short:        "Configuration",
		SilenceUsage: true,
	}
	return cmd
}
