package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/virtualmachineimagebuilder/mgmt/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppBuilderCancel struct {
	*AppBuilder
	Input string
}

func (app *AppBuilder) AppBuilderCancelCmder() cmder.Cmder {
	return &AppBuilderCancel{AppBuilder: app}
}

func (app *AppBuilderCancel) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cancel",
		Short:        "Cancel image builder",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func (app *AppBuilderCancel) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	templateFuture, err := templatesClient.Cancel(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}
	err = templateFuture.WaitForCompletionRef(ctx, templatesClient.Client)
	if err != nil {
		return err
	}

	return nil
}
