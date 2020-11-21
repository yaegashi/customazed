package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppMachine struct {
	*App
}

func (app *App) AppMachineCmder() cmder.Cmder {
	return &AppMachine{App: app}
}

func (app *AppMachine) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "machine",
		Short:        "Azure VM Custom Script Extension",
		SilenceUsage: true,
	}
	return cmd
}
