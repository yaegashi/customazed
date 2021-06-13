package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/virtualmachineimagebuilder/mgmt/2020-02-14/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderRun is app builder run command
type AppBuilderRun struct {
	*AppBuilder
}

// AppBuilderRunCmder returns Cmder for app builder run
func (app *AppBuilder) AppBuilderRunCmder() cmder.Cmder {
	return &AppBuilderRun{AppBuilder: app}
}

// Cmd returns Command for app builder run
func (app *AppBuilderRun) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "Run image build",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app builder run
func (app *AppBuilderRun) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.LogBuilderName()
	app.Log("Running image build...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	_, err = templatesClient.Run(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	app.Logf("Image build started")
	app.Logf("To show the image build status:    customazed builder show-status")
	app.Logf("To watch customization.log output: customazed builder show-logs -F")

	return nil
}
