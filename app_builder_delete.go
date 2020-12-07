package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/virtualmachineimagebuilder/mgmt/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppBuilderDelete struct {
	*AppBuilder
}

func (app *AppBuilder) AppBuilderDeleteCmder() cmder.Cmder {
	return &AppBuilderDelete{AppBuilder: app}
}

func (app *AppBuilderDelete) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete image template",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func (app *AppBuilderDelete) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.Log("Deleting image template...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	templateFuture, err := templatesClient.Delete(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}
	err = templateFuture.WaitForCompletionRef(ctx, templatesClient.Client)
	if err != nil {
		return err
	}

	return nil
}
