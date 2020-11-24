package main

import (
	"context"
	"net/url"
	"os"

	"github.com/yaegashi/customazed/utils/ssutil"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

func (app *App) StorageAccount(ctx context.Context) (*storage.Account, error) {
	if app._StorageAccount == nil {
		err := app.StorageGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._StorageAccount, nil
}

func (app *App) StorageContainer(ctx context.Context) (*storage.BlobContainer, error) {
	if app._StorageContainer == nil {
		err := app.StorageGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._StorageContainer, nil
}

func (app *App) StorageGet(ctx context.Context) error {
	cfgStorage := app.Config.Storage
	if ssutil.HasEmpty(cfgStorage.Location, cfgStorage.ResourceGroup, cfgStorage.AccountName, cfgStorage.ContainerName) {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	accountsClient := storage.NewAccountsClient(app.Config.SubscriptionID)
	accountsClient.Authorizer = authorizer
	account, err := accountsClient.GetProperties(ctx, app.Config.Storage.ResourceGroup, app.Config.Storage.AccountName, "")
	if err != nil {
		return err
	}

	containerClient := storage.NewBlobContainersClient(app.Config.SubscriptionID)
	containerClient.Authorizer = authorizer
	container, err := containerClient.Get(ctx, app.Config.Storage.ResourceGroup, app.Config.Storage.AccountName, app.Config.Storage.ContainerName)
	if err != nil {
		return err
	}

	app._StorageAccount = &account
	app._StorageContainer = &container

	app.Config.Storage.AccountID = *account.ID
	app.Config.Storage.ContainerID = *container.ID

	return nil
}

func (app *App) StorageSetup(ctx context.Context) error {
	cfgStorage := app.Config.Storage
	if ssutil.HasEmpty(cfgStorage.Location, cfgStorage.ResourceGroup, cfgStorage.AccountName, cfgStorage.ContainerName) {
		app.Log("Storage: missing configuration")
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.Logf("Storage: creating resource group: %s", app.Config.Storage.ResourceGroup)
	groupsClient := resources.NewGroupsClient(app.Config.SubscriptionID)
	groupsClient.Authorizer = authorizer
	groupsParams := resources.Group{
		Location: &app.Config.Storage.Location,
	}
	_, err = groupsClient.CreateOrUpdate(ctx, app.Config.Storage.ResourceGroup, groupsParams)
	if err != nil {
		return err
	}

	app.Logf("Storage: creating storage account: %s", app.Config.Storage.AccountName)
	accountsClient := storage.NewAccountsClient(app.Config.SubscriptionID)
	accountsClient.Authorizer = authorizer
	accountsParams := storage.AccountCreateParameters{
		Location: &app.Config.Storage.Location,
		Kind:     "StorageV2",
		Sku:      &storage.Sku{Name: "Standard_LRS"},
	}
	accountFuture, err := accountsClient.Create(ctx, app.Config.Storage.ResourceGroup, app.Config.Storage.AccountName, accountsParams)
	if err != nil {
		return err
	}
	err = accountFuture.WaitForCompletionRef(ctx, accountsClient.Client)
	if err != nil {
		return err
	}

	app.Logf("Storage: creating blob container: %s", app.Config.Storage.ContainerName)
	containerClient := storage.NewBlobContainersClient(app.Config.SubscriptionID)
	containerClient.Authorizer = authorizer
	container := storage.BlobContainer{
		ContainerProperties: &storage.ContainerProperties{
			PublicAccess: storage.PublicAccessNone,
		},
	}
	container, err = containerClient.Create(ctx, app.Config.Storage.ResourceGroup, app.Config.Storage.AccountName, app.Config.Storage.ContainerName, container)
	if err != nil {
		return err
	}

	return app.StorageGet(ctx)
}

func (app *App) StorageUpload(ctx context.Context, path string) (string, error) {
	uploadMap := app.StorageMap()
	if _, ok := uploadMap[path]; ok {
		return "", nil
	}

	token, err := app.StorageToken()
	if err != nil {
		return "", err
	}

	account, err := app.StorageAccount(ctx)
	if err != nil {
		return "", err
	}

	app.Logf("Uploading %s...", path)

	r, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	p := azblob.NewPipeline(azblob.NewTokenCredential(token.OAuthToken(), nil), azblob.PipelineOptions{})
	u, _ := url.Parse(*account.PrimaryEndpoints.Blob)
	storageURL := azblob.NewServiceURL(*u, p)
	containerURL := storageURL.NewContainerURL(app.Config.Storage.ContainerName)
	blobURL := containerURL.NewBlockBlobURL(path)
	_, err = azblob.UploadFileToBlockBlob(ctx, r, blobURL, azblob.UploadToBlockBlobOptions{})
	if err != nil {
		return "", err
	}

	uploadMap[path] = blobURL.String()

	return blobURL.String(), nil
}

func (app *App) StorageMap() map[string]string {
	if app._StorageMap == nil {
		app._StorageMap = make(map[string]string)
	}
	return app._StorageMap
}
