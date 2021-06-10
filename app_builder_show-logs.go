package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-04-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilderShowLogs is app builder show command
type AppBuilderShowLogs struct {
	*AppBuilder
	LogName string
	Tail    int64
	Follow  bool
}

// AppBuilderShowLogsCmder returns Cmder for app builder show
func (app *AppBuilder) AppBuilderShowLogsCmder() cmder.Cmder {
	return &AppBuilderShowLogs{AppBuilder: app}
}

// Cmd returns Command for app builder show
func (app *AppBuilderShowLogs) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show-logs",
		Short:        "Show image build logs",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	cmd.PersistentFlags().StringVarP(&app.LogName, "log", "", "", "name of the log to show (default the latest log)")
	cmd.PersistentFlags().Int64VarP(&app.Tail, "tail", "", 1024, "last bytes of the log to show (0 means all)")
	cmd.PersistentFlags().BoolVarP(&app.Follow, "follow", "F", false, "follow log output")
	return cmd
}

// RunE is main routine for app builder show
func (app *AppBuilderShowLogs) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.LogBuilderName()

	// Find the resource group with specified tags
	groupsClinet := resources.NewGroupsClient(app.Config.SubscriptionID)
	groupsClinet.Authorizer = authorizer
	groups, err := groupsClinet.ListComplete(ctx, "tagName eq 'createdBy' and tagValue eq 'AzureVMImageBuilder'", nil)
	if err != nil {
		return err
	}
	var group *resources.Group
	for groups.NotDone() {
		g := groups.Value()
		if name, ok := g.Tags["imageTemplateName"]; ok {
			if resourceGroupName := g.Tags["imageTemplateResourceGroupName"]; ok {
				if strings.EqualFold(*name, app.Config.Builder.BuilderName) && strings.EqualFold(*resourceGroupName, app.Config.Builder.ResourceGroup) {
					group = &g
					break
				}
			}
		}
		err := groups.NextWithContext(ctx)
		if err != nil {
			return err
		}
	}
	if group == nil {
		return fmt.Errorf("builder resource group not found")
	}
	app.Logf("Builder resource group: %s", *group.Name)

	// Find the storage account with specified tags
	accountsClient := storage.NewAccountsClient(app.Config.SubscriptionID)
	accountsClient.Authorizer = authorizer
	accounts, err := accountsClient.ListByResourceGroupComplete(ctx, *group.Name)
	if err != nil {
		return err
	}
	var account *storage.Account
	for accounts.NotDone() {
		a := accounts.Value()
		if createdby, ok := a.Tags["createdby"]; ok {
			if *createdby == "azureimagebuilder" {
				account = &a
				break
			}
		}
		err := accounts.NextWithContext(ctx)
		if err != nil {
			return err
		}
	}
	if account == nil {
		return fmt.Errorf("builder storage account not found")
	}
	app.Logf("Builder storage account: %s", *account.Name)

	// Get shared access keys
	keyResult, err := accountsClient.ListKeys(ctx, *group.Name, *account.Name, "")
	if err != nil {
		return err
	}
	accountKeys := *keyResult.Keys

	// Find packerlogs container URL
	credential, err := azblob.NewSharedKeyCredential(*account.Name, *accountKeys[0].Value)
	if err != nil {
		return err
	}
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	endpointURL, _ := url.Parse(*account.PrimaryEndpoints.Blob)
	serviceURL := azblob.NewServiceURL(*endpointURL, pipeline)
	containerURL := serviceURL.NewContainerURL("packerlogs")

	// Enumerate log blobs in packerlogs container
	var blobItems []*azblob.BlobItemInternal
	for marker := (azblob.Marker{}); marker.NotDone(); {
		blobRes, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return err
		}
		for i := range blobRes.Segment.BlobItems {
			blobItems = append(blobItems, &blobRes.Segment.BlobItems[i])
		}
		marker = blobRes.NextMarker
	}
	if len(blobItems) == 0 {
		return fmt.Errorf("no log blob found in packerlogs container")
	}

	// Sort log blobs by creation time
	sort.Slice(blobItems, func(i, j int) bool {
		var a, b time.Time
		if blobItems[i].Properties.CreationTime != nil {
			a = *blobItems[i].Properties.CreationTime
		}
		if blobItems[j].Properties.CreationTime != nil {
			b = *blobItems[j].Properties.CreationTime
		}
		return a.Before(b)
	})

	// Select log blob to show
	var blobItem *azblob.BlobItemInternal
	name := app.LogName
	if name == "" {
		name = blobItems[len(blobItems)-1].Name
	}
	app.Logf("Enumerating log blobs...")
	for _, item := range blobItems {
		selected := ""
		if item.Name == name {
			blobItem = item
			selected = " (selected)"
		}
		app.Logf("  %s %s%s", item.Properties.CreationTime.Format(time.RFC3339), item.Name, selected)
	}

	// Calc offset
	var offset int64
	tail := app.Tail
	if tail < 0 {
		tail = 0
	}
	if tail > 0 {
		offset = *blobItem.Properties.ContentLength - tail
	}
	if offset < 0 {
		offset = 0
	}

	// Download log blob and dump it to stdout
	blobURL := containerURL.NewAppendBlobURL(blobItem.Name)
	app.Logf("Getting %s", blobURL)
	if offset > 0 {
		os.Stdout.WriteString("....")
	}
	for {
		res, err := blobURL.Download(ctx, offset, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
		if err == nil {
			n, err := io.Copy(os.Stdout, res.Body(azblob.RetryReaderOptions{}))
			res.Response().Body.Close()
			if err != nil {
				return err
			}
			offset += n
		} else {
			storageErr, ok := err.(azblob.StorageError)
			if !ok || storageErr.Response().StatusCode != http.StatusRequestedRangeNotSatisfiable {
				return err
			}
		}
		if !app.Follow {
			break
		}
		time.Sleep(2 * time.Second)
	}

	return nil
}
