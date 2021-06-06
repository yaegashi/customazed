package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/virtualmachineimagebuilder/mgmt/2020-02-14/virtualmachineimagebuilder"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderShowStatus is app builder show-status command
type AppBuilderShowStatus struct {
	*AppBuilder
}

// AppBuilderShowStatusCmder returns Cmder for app builder show-status
func (app *AppBuilder) AppBuilderShowStatusCmder() cmder.Cmder {
	return &AppBuilderShowStatus{AppBuilder: app}
}

// Cmd returns Command for app builder show-status
func (app *AppBuilderShowStatus) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show-status",
		Short:        "Show image template status",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app builder show-status
func (app *AppBuilderShowStatus) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.LogBuilderName()
	app.Log("Getting image template status...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	template, err := templatesClient.Get(ctx, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	m := map[string]interface{}{}
	m["provisioningState"] = template.ProvisioningState
	if template.LastRunStatus != nil {
		m["lastRunStatus"] = template.LastRunStatus
	}
	if template.ProvisioningError != nil {
		m["provisioningError"] = template.ProvisioningError
	}

	app.Dump(m)

	return nil
}
