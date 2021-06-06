package main

import (
	"context"
	"fmt"

	"github.com/yaegashi/customazed/utils/ssutil"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-03-01/compute"
)

func (app *App) Image(ctx context.Context) (*compute.Image, error) {
	if app._Image == nil {
		err := app.ImageGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._Image, nil
}

func (app *App) ImageValid() bool {
	cfg := app.Config.Image
	if ssutil.HasEmpty(cfg.Location, cfg.ResourceGroup, cfg.ImageName) {
		app.Log("Image: missing configuration")
		return false
	}
	return true
}

func (app *App) ImageGet(ctx context.Context) error {
	if !app.ImageValid() {
		return nil
	}

	app._Image = &compute.Image{}
	app.Config.Image.ImageID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/images/%s", app.Config.SubscriptionID, app.Config.Image.ResourceGroup, app.Config.Image.ImageName)

	return nil
}

func (app *App) ImageSetup(ctx context.Context) error {
	if !app.ImageValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.Logf("Image: creating resource group: %s", app.Config.Image.ResourceGroup)
	groupsClient := resources.NewGroupsClient(app.Config.SubscriptionID)
	groupsClient.Authorizer = authorizer
	groupsParams := resources.Group{
		Location: &app.Config.Image.Location,
	}
	_, err = groupsClient.CreateOrUpdate(ctx, app.Config.Image.ResourceGroup, groupsParams)
	if err != nil {
		return err
	}

	return nil
}
