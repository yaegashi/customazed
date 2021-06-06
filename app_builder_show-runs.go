package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/virtualmachineimagebuilder/mgmt/2020-02-14/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderShowRuns is app builder show-runs command
type AppBuilderShowRuns struct {
	*AppBuilder
}

// AppBuilderShowRunsCmder returns Cmder for app builder show-runs
func (app *AppBuilder) AppBuilderShowRunsCmder() cmder.Cmder {
	return &AppBuilderShowRuns{AppBuilder: app}
}

// Cmd returns Command for app builder show-runs
func (app *AppBuilderShowRuns) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show-runs",
		Short:        "Show image template run outputs",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app builder show-runs
func (app *AppBuilderShowRuns) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.LogBuilderName()
	app.Log("Getting image template run outputs...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	result, err := templatesClient.ListRunOutputsComplete(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	type runOutput map[string]interface{}
	var runOutputs []runOutput
	for result.NotDone() {
		r := result.Value()
		m := runOutput{}
		m["name"] = r.Name
		m["provisioningState"] = r.ProvisioningState
		if r.ArtifactID != nil {
			m["artifactId"] = r.ArtifactID
		}
		if r.ArtifactURI != nil {
			m["artifactUri"] = r.ArtifactURI
		}
		runOutputs = append(runOutputs, m)
		err := result.Next()
		if err != nil {
			return err
		}
	}

	app.Dump(runOutputs)

	return nil
}
