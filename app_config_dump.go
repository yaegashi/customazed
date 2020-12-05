package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppConfigDump struct {
	*AppConfig
}

func (app *AppConfig) AppConfigDumpCmder() cmder.Cmder {
	return &AppConfigDump{AppConfig: app}
}

func (app *AppConfigDump) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dump",
		Short:        "Dump configuration",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func (app *AppConfigDump) RunE(cmder *cobra.Command, args []string) error {
	app.Log("Dumping config")
	app.Dump(app.Config)
	return nil
}
