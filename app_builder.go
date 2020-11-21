package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppBuilder struct {
	*App
}

func (app *App) AppBuilderCmder() cmder.Cmder {
	return &AppBuilder{App: app}
}

func (app *AppBuilder) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "builder",
		Short:        "Azure VM Image Builder",
		SilenceUsage: true,
	}
	return cmd
}
