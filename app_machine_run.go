package main

import (
	"context"
	"net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppMachineRun struct {
	*AppMachine
}

func (app *AppMachine) AppMachineRunCmder() cmder.Cmder {
	return &AppMachineRun{AppMachine: app}
}

func (app *AppMachineRun) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "run VM extension",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func (app *AppMachineRun) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	storageToken, err := app.StorageToken()
	if err != nil {
		return err
	}

	storageAccount, err := app.StorageAccount(ctx)
	if err != nil {
		return err
	}

	p := azblob.NewPipeline(azblob.NewTokenCredential(storageToken.OAuthToken(), nil), azblob.PipelineOptions{})
	u, _ := url.Parse(*storageAccount.PrimaryEndpoints.Blob)
	storageURL := azblob.NewServiceURL(*u, p)
	containerURL := storageURL.NewContainerURL(app.Config.Storage.ContainerName)

	var fileURIs []string
	for _, file := range app.Config.Machine.Files {
		app.Logf("Uploading %s", file)
		r, err := os.Open(file)
		if err != nil {
			return err
		}
		defer r.Close()
		blobURL := containerURL.NewBlockBlobURL(file)
		_, err = azblob.UploadFileToBlockBlob(ctx, r, blobURL, azblob.UploadToBlockBlobOptions{})
		if err != nil {
			return err
		}
		fileURIs = append(fileURIs, blobURL.String())
	}

	extensionParams := NewWindowsCustomScriptExtension(app.Config.Machine.Location)
	extensionParams.ProtectedSettings = map[string]interface{}{
		"fileURIs":         fileURIs,
		"commandToExecute": app.Config.Machine.Command,
		"managedIdentity":  map[string]string{},
	}

	app.Log("Executing VM extension...")
	extensionsClient := compute.NewVirtualMachineExtensionsClient(app.Config.SubscriptionID)
	extensionsClient.Authorizer = authorizer
	extensionFuture, err := extensionsClient.CreateOrUpdate(ctx, app.Config.Machine.ResourceGroup, app.Config.Machine.MachineName, "CustomScriptExtension", *extensionParams)
	if err != nil {
		return err
	}
	err = extensionFuture.WaitForCompletionRef(ctx, extensionsClient.Client)
	if err != nil {
		app.Logf("Execution failed: run \"%s vm status\" for the detail", cmd.Root().Name())
	}

	return err
}
