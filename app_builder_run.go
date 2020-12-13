package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/virtualmachineimagebuilder/mgmt/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderRun is app builder run command
type AppBuilderRun struct {
	*AppBuilder
	Input string
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
	templateFuture, err := templatesClient.Run(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}
	err = templateFuture.WaitForCompletionRef(ctx, templatesClient.Client)
	if err != nil {
		return err
	}

	return nil
}
