package main

import (
	"context"

	"github.com/yaegashi/customazed/utils/ssutil"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
)

func NewWindowsCustomScriptExtension(location string) *compute.VirtualMachineExtension {
	return &compute.VirtualMachineExtension{
		Location: &location,
		VirtualMachineExtensionProperties: &compute.VirtualMachineExtensionProperties{
			Publisher:               to.StringPtr("Microsoft.Compute"),
			Type:                    to.StringPtr("CustomScriptExtension"),
			TypeHandlerVersion:      to.StringPtr("1.10"),
			AutoUpgradeMinorVersion: to.BoolPtr(true),
		},
	}
}

func NewLinuxCustomScriptExtension(location string) *compute.VirtualMachineExtension {
	return &compute.VirtualMachineExtension{
		Location: &location,
		VirtualMachineExtensionProperties: &compute.VirtualMachineExtensionProperties{
			Publisher:               to.StringPtr("Microsoft.Azure.Extensions"),
			Type:                    to.StringPtr("CustomScript"),
			TypeHandlerVersion:      to.StringPtr("2.1"),
			AutoUpgradeMinorVersion: to.BoolPtr(true),
		},
	}
}

func (app *App) Machine(ctx context.Context) (*compute.VirtualMachine, error) {
	if app._Machine == nil {
		err := app.MachineGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._Machine, nil
}

func (app *App) MachineValid() bool {
	cfg := app.Config.Machine
	if ssutil.HasEmpty(cfg.ResourceGroup, cfg.MachineName) {
		app.Log("Machine: missing configuration")
		return false
	}
	return true
}

func (app *App) MachineGet(ctx context.Context) error {
	if !app.MachineValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	machinesClient := compute.NewVirtualMachinesClient(app.Config.SubscriptionID)
	machinesClient.Authorizer = authorizer
	machine, err := machinesClient.Get(ctx, app.Config.Machine.ResourceGroup, app.Config.Machine.MachineName, "")
	if err != nil {
		return err
	}

	app._Machine = &machine
	app.Config.Machine.MachineID = *machine.ID

	return nil
}

func (app *App) MachineSetup(ctx context.Context) error {
	if !app.MachineValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.Log("Machine: enabling system assigned identity")
	machinesClient := compute.NewVirtualMachinesClient(app.Config.SubscriptionID)
	machinesClient.Authorizer = authorizer
	machineUpdate := compute.VirtualMachineUpdate{
		Identity: &compute.VirtualMachineIdentity{
			Type: compute.ResourceIdentityTypeSystemAssigned,
		},
	}
	machineFuture, err := machinesClient.Update(ctx, app.Config.Machine.ResourceGroup, app.Config.Machine.MachineName, machineUpdate)
	if err != nil {
		return err
	}
	err = machineFuture.WaitForCompletionRef(ctx, machinesClient.Client)
	if err != nil {
		return err
	}

	return app.MachineGet(ctx)
}
