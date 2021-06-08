package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/virtualmachineimagebuilder/mgmt/2020-02-14/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderCancel is app builder cancel command
type AppBuilderCancel struct {
	*AppBuilder
}

// AppBuilderCancelCmder returns Cmder for app builder cancel
func (app *AppBuilder) AppBuilderCancelCmder() cmder.Cmder {
	return &AppBuilderCancel{AppBuilder: app}
}

// Cmd returns Cmd for app builder cancel
func (app *AppBuilderCancel) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cancel",
		Short:        "Cancel image build",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app builder cancel
func (app *AppBuilderCancel) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.LogBuilderName()
	app.Log("Canceling image build...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	templateFuture, err := templatesClient.Cancel(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	return app.WaitForCompletion(ctx, &templateFuture, templatesClient.Client)
}
