package main

import (
	"context"
	"errors"
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

func (app *App) StorageValid() bool {
	cfg := app.Config.Storage
	if ssutil.HasEmpty(cfg.Location, cfg.ResourceGroup, cfg.AccountName, cfg.ContainerName) {
		app.Log("Storage: missing configuration")
		return false
	}
	return true
}

func (app *App) StorageGet(ctx context.Context) error {
	if !app.StorageValid() {
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
	if !app.StorageValid() {
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

type StorageUploader interface {
	Valid() bool
	Files() int
	Add(s string) (string, error)
	Execute(ctx context.Context) error
}

type DisabledStorageUploader string

func (d DisabledStorageUploader) Valid() bool                       { return false }
func (d DisabledStorageUploader) Files() int                        { return 0 }
func (d DisabledStorageUploader) Add(s string) (string, error)      { return "", errors.New(string(d)) }
func (d DisabledStorageUploader) Execute(ctx context.Context) error { return errors.New(string(d)) }

type BlobStorageUploader struct {
	app          *App
	prefix       string
	uploadMap    map[string]string
	containerURL azblob.ContainerURL
	valid        bool
}

func (app *App) NewStorageUploader(ctx context.Context) StorageUploader {
	if !app.StorageValid() {
		return DisabledStorageUploader("upload: no storage configuration")
	}
	endpoint := "https://" + app.Config.Storage.AccountName
	valid := !app.NoLogin
	if valid {
		account, err := app.StorageAccount(ctx)
		if err == nil {
			endpoint = *account.PrimaryEndpoints.Blob
		} else {
			valid = false
		}
	}
	endpointURL, _ := url.Parse(endpoint)
	serviceURL := azblob.NewServiceURL(*endpointURL, nil)
	containerURL := serviceURL.NewContainerURL(app.Config.Storage.ContainerName)
	return &BlobStorageUploader{
		app:          app,
		prefix:       app.HashID("upload"),
		uploadMap:    map[string]string{},
		containerURL: containerURL,
		valid:        valid,
	}
}

func (su *BlobStorageUploader) path(path string) string { return su.prefix + "/" + path }
func (su *BlobStorageUploader) Valid() bool             { return su.valid }
func (su *BlobStorageUploader) Files() int              { return len(su.uploadMap) }

func (su *BlobStorageUploader) Add(path string) (string, error) {
	stringBlobURL, ok := su.uploadMap[path]
	if !ok {
		su.app.Logf("Blob: adding %s", path)
		stringBlobURL = su.containerURL.NewBlockBlobURL(su.path(path)).String()
		su.uploadMap[path] = stringBlobURL
	}
	return stringBlobURL, nil
}

func (su *BlobStorageUploader) Execute(ctx context.Context) error {
	if !su.valid {
		return nil
	}

	app := su.app

	token, err := app.StorageToken()
	if err != nil {
		return err
	}

	account, err := app.StorageAccount(ctx)
	if err != nil {
		return err
	}

	p := azblob.NewPipeline(azblob.NewTokenCredential(token.OAuthToken(), nil), azblob.PipelineOptions{})
	endpointURL, _ := url.Parse(*account.PrimaryEndpoints.Blob)
	serviceURL := azblob.NewServiceURL(*endpointURL, p)
	containerURL := serviceURL.NewContainerURL(app.Config.Storage.ContainerName)
	stringPrefixURL := containerURL.NewBlockBlobURL(su.prefix).String()
	app.Logf("Blob: destination %s", stringPrefixURL)

	for path := range su.uploadMap {
		app.Logf("Blob: uploading %s", path)
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		blobURL := containerURL.NewBlockBlobURL(su.path(path))
		_, err = azblob.UploadFileToBlockBlob(ctx, r, blobURL, azblob.UploadToBlockBlobOptions{})
		r.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
