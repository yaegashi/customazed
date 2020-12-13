package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppMachine is app machine command
type AppMachine struct {
	*App
}

// AppMachineCmder returns Cmder for app machine
func (app *App) AppMachineCmder() cmder.Cmder {
	return &AppMachine{App: app}
}

// Cmd returns Command for app machine
func (app *AppMachine) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "machine",
		Short:        "Azure VM Custom Script Extension",
		SilenceUsage: true,
	}
	return cmd
}
