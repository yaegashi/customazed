package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppConfigDump is app config dump command
type AppConfigDump struct {
	*AppConfig
}

// AppConfigDumpCmder returns Cmder for app config dump
func (app *AppConfig) AppConfigDumpCmder() cmder.Cmder {
	return &AppConfigDump{AppConfig: app}
}

// Cmd returns Command for app config dump
func (app *AppConfigDump) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dump",
		Short:        "Dump configuration",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app config dump
func (app *AppConfigDump) RunE(cmder *cobra.Command, args []string) error {
	app.Log("Dumping configuration")
	app.Dump(app.Config)
	return nil
}
