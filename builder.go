package main

import (
	"context"
	"fmt"

	"github.com/yaegashi/customazed/utils/ssutil"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/services/virtualmachineimagebuilder/mgmt/2020-02-14/virtualmachineimagebuilder"
)

func (app *App) Builder(ctx context.Context) (*virtualmachineimagebuilder.ImageTemplate, error) {
	if app._Builder == nil {
		err := app.BuilderGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._Builder, nil
}

func (app *App) BuilderValid() bool {
	cfg := app.Config.Builder
	if ssutil.HasEmpty(cfg.Location, cfg.ResourceGroup, cfg.BuilderName) {
		app.Log("Builder: missing configuration")
		return false
	}
	return true
}

func (app *App) BuilderGet(ctx context.Context) error {
	if !app.BuilderValid() {
		return nil
	}

	app._Builder = &virtualmachineimagebuilder.ImageTemplate{}
	app.Config.Builder.BuilderID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.VirtualMachineImages/imageTemplates/%s", app.Config.SubscriptionID, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)

	return nil
}

func (app *App) BuilderSetup(ctx context.Context) error {
	if !app.BuilderValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.Logf("Builder: creating resource group: %s", app.Config.Builder.ResourceGroup)
	groupsClient := resources.NewGroupsClient(app.Config.SubscriptionID)
	groupsClient.Authorizer = authorizer
	groupsParams := resources.Group{
		Location: &app.Config.Builder.Location,
	}
	_, err = groupsClient.CreateOrUpdate(ctx, app.Config.Builder.ResourceGroup, groupsParams)
	if err != nil {
		return err
	}

	return nil
}
