package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/virtualmachineimagebuilder/mgmt/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppBuilderShow struct {
	*AppBuilder
}

func (app *AppBuilder) AppBuilderShowCmder() cmder.Cmder {
	return &AppBuilderShow{AppBuilder: app}
}

func (app *AppBuilderShow) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show",
		Short:        "Show image template",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func (app *AppBuilderShow) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	template, err := templatesClient.Get(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	app.Dump(template)

	return nil
}
