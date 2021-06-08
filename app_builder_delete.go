package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/virtualmachineimagebuilder/mgmt/2020-02-14/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderDelete is app builder delete command
type AppBuilderDelete struct {
	*AppBuilder
}

// AppBuilderDeleteCmder returns Cmder for app builder delete
func (app *AppBuilder) AppBuilderDeleteCmder() cmder.Cmder {
	return &AppBuilderDelete{AppBuilder: app}
}

// Cmd returns Command for app builder delete
func (app *AppBuilderDelete) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete image template",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app builder delete
func (app *AppBuilderDelete) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.LogBuilderName()
	app.Log("Deleting image template...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	templateFuture, err := templatesClient.Delete(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	return app.WaitForCompletion(ctx, &templateFuture, templatesClient.Client)
}
