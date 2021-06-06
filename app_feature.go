package main

import (
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

var appFeatureFeatures = [][]string{
	{"Microsoft.VirtualMachineImages", "VirtualMachineTemplatePreview"},
}

var appFeatureProviders = []string{
	"Microsoft.VirtualMachineImages",
	"Microsoft.KeyVault",
	"Microsoft.Compute",
	"Microsoft.Network",
	"Microsoft.Storage",
}

// AppFeature is app feature command
type AppFeature struct {
	*App
}

// AppFeatureCmder returns Cmder for app feature
func (app *App) AppFeatureCmder() cmder.Cmder {
	return &AppFeature{App: app}
}

// Cmd returns Command for app feature
func (app *AppFeature) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "feature",
		Short:        "Manage Azure features/providers",
		SilenceUsage: true,
	}
	return cmd
}
