package main

import (
	"context"
	"errors"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppMachineShowStatus is app machine show-status command
type AppMachineShowStatus struct {
	*AppMachine
}

// AppMachineShowStatusCmder returns Cmder for app machine show-status
func (app *AppMachine) AppMachineShowStatusCmder() cmder.Cmder {
	return &AppMachineShowStatus{AppMachine: app}
}

// Cmd returns Cmder for app machine show-status
func (app *AppMachineShowStatus) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show-status",
		Short:        "show last status of VM extension",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app machine show-status
func (app *AppMachineShowStatus) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	extensionsClient := compute.NewVirtualMachineExtensionsClient(app.Config.SubscriptionID)
	extensionsClient.Authorizer = authorizer
	result, err := extensionsClient.Get(ctx, app.Config.Machine.ResourceGroup, app.Config.Machine.MachineName, "CustomScriptExtension", "instanceView")
	if err != nil {
		return err
	}
	if result.VirtualMachineExtensionProperties.InstanceView == nil {
		return errors.New("missing extension instance view (maybe virtual machine is not running)")
	}
	for _, status := range *result.VirtualMachineExtensionProperties.InstanceView.Statuses {
		app.Logf("%s: %s\n%s", *status.Code, *status.DisplayStatus, *status.Message)
	}
	if result.VirtualMachineExtensionProperties.InstanceView.Substatuses != nil {
		for _, status := range *result.VirtualMachineExtensionProperties.InstanceView.Substatuses {
			code := strings.Split(*status.Code, "/")
			if len(code) == 3 && code[0] == "ComponentStatus" {
				msg := *status.Message
				if len(msg) > 0 && msg[len(msg)-1] != '\n' {
					msg = msg + "\n"
				}
				app.Logf("%s:\n%s", code[1], msg)
			}
		}
	}

	return nil
}
