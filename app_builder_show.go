package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/virtualmachineimagebuilder/mgmt/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderShow is app builder show command
type AppBuilderShow struct {
	*AppBuilder
}

// AppBuilderShowCmder returns Cmder for app builder show
func (app *AppBuilder) AppBuilderShowCmder() cmder.Cmder {
	return &AppBuilderShow{AppBuilder: app}
}

// Cmd returns Command for app builder show
func (app *AppBuilderShow) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show",
		Short:        "Show image template",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app builder show
func (app *AppBuilderShow) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.LogBuilderName()
	app.Log("Getting image template...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	template, err := templatesClient.Get(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	app.Dump(template)

	return nil
}
