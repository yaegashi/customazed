package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-03-01/compute"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppMachineRun is app machine run command
type AppMachineRun struct {
	*AppMachine
	Input string
}

// AppMachineRunCmder returns Cmder for app machine run
func (app *AppMachine) AppMachineRunCmder() cmder.Cmder {
	return &AppMachineRun{AppMachine: app}
}

// Cmd returns Command for app machine rune
func (app *AppMachineRun) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "run VM extension",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&app.Input, "input", "i", "customazed_machine.json", "input file path")
	return cmd
}

// CustomScriptSettings is input object of app machine run
type CustomScriptSettings struct {
	FileUris         []string `json:"fileUris,omitempty"`
	CommandToExecute string   `json:"commandToExecute,omitempty"`
	SkipDos2Unix     bool     `json:"skipDos2Unix,omitempty"`
	Timestamp        int      `json:"timestamp,omitempty"`
}

// RunE is main routine for app machine run
func (app *AppMachineRun) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	machine, err := app.Machine(ctx)
	if err != nil {
		return err
	}

	app.Logf("Loading custom script settings %s", app.Input)
	var b []byte
	if app.Input == "-" {
		b, err = ioutil.ReadAll(os.Stdin)
	} else {
		b, err = ioutil.ReadFile(app.Input)
	}
	if err != nil {
		return err
	}

	var settings *CustomScriptSettings
	err = json.Unmarshal(b, &settings)
	if err != nil {
		return err
	}

	if settings.Timestamp == 0 {
		settings.Timestamp = int(time.Now().Unix())
	}

	su := app.NewStorageUploader(ctx)
	tv := app.NewTemplateVariable(su)
	err = tv.Resolve(settings)
	if err != nil {
		return err
	}

	var extensionParams *compute.VirtualMachineExtension
	switch machine.StorageProfile.OsDisk.OsType {
	case "Windows":
		extensionParams = NewWindowsCustomScriptExtension(*machine.Location)
		extensionParams.ProtectedSettings = map[string]interface{}{
			"fileUris":         settings.FileUris,
			"commandToExecute": settings.CommandToExecute,
			"timestamp":        settings.Timestamp,
			"managedIdentity":  map[string]string{},
		}
	case "Linux":
		extensionParams = NewLinuxCustomScriptExtension(*machine.Location)
		extensionParams.Settings = map[string]interface{}{
			"skipDos2Unix": settings.SkipDos2Unix,
			"timestamp":    settings.Timestamp,
		}
		extensionParams.ProtectedSettings = map[string]interface{}{
			"fileUris":         settings.FileUris,
			"commandToExecute": settings.CommandToExecute,
			"managedIdentity":  map[string]string{},
		}
	default:
		return fmt.Errorf("VM has unknown OS type: %s", machine.StorageProfile.OsDisk.OsType)
	}

	app.Dump(settings)
	app.Prompt("Files to upload: %d", su.Files())

	if su.Valid() && su.Files() > 0 {
		err = su.Execute(ctx)
		if err != nil {
			return err
		}
	}

	app.Log("Executing VM extension...")
	extensionsClient := compute.NewVirtualMachineExtensionsClient(app.Config.SubscriptionID)
	extensionsClient.Authorizer = authorizer
	extensionFuture, err := extensionsClient.CreateOrUpdate(ctx, app.Config.Machine.ResourceGroup, app.Config.Machine.MachineName, "CustomScriptExtension", *extensionParams)
	if err != nil {
		return err
	}
	err = extensionFuture.WaitForCompletionRef(ctx, extensionsClient.Client)
	result := "Success"
	if err != nil {
		result = "Failure"
	}
	app.Logf("%s: use \"%s machine show-status\" to see the output", result, cmd.Root().Name())

	return err
}
