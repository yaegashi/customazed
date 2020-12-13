package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilder is app builder command
type AppBuilder struct {
	*App
}

// AppBuilderCmder returns Cmder for app builder
func (app *App) AppBuilderCmder() cmder.Cmder {
	return &AppBuilder{App: app}
}

// Cmd returns Command for app builder
func (app *AppBuilder) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "builder",
		Short:        "Azure VM Image Builder",
		SilenceUsage: true,
	}
	return cmd
}

// LogBuilderName shows current builder name
func (app *AppBuilder) LogBuilderName() {
	app.Logf("Current builder name: %s", app.Config.Builder.BuilderName)
}
