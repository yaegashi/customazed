package main

import (
	"context"

	"github.com/yaegashi/customazed/utils/ssutil"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/services/msi/mgmt/2018-11-30/msi"
)

func (app *App) Identity(ctx context.Context) (*msi.Identity, error) {
	if app._Identity == nil {
		err := app.IdentityGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._Identity, nil
}

func (app *App) IdentityValid() bool {
	cfg := app.Config.Identity
	if ssutil.HasEmpty(cfg.Location, cfg.ResourceGroup, cfg.IdentityName) {
		app.Log("Identity: missing configuration")
		return false
	}
	return true
}

func (app *App) IdentityGet(ctx context.Context) error {
	if !app.IdentityValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	msiClient := msi.NewUserAssignedIdentitiesClient(app.Config.SubscriptionID)
	msiClient.Authorizer = authorizer
	identity, err := msiClient.Get(ctx, app.Config.Identity.ResourceGroup, app.Config.Identity.IdentityName)
	if err != nil {
		return err
	}

	app._Identity = &identity
	app.Config.Identity.IdentityID = *identity.ID

	return nil
}

func (app *App) IdentitySetup(ctx context.Context) error {
	if !app.IdentityValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.Logf("Identity: creating resource group: %s", app.Config.Identity.ResourceGroup)
	groupsClient := resources.NewGroupsClient(app.Config.SubscriptionID)
	groupsClient.Authorizer = authorizer
	groupsParams := resources.Group{
		Location: &app.Config.Identity.Location,
	}
	_, err = groupsClient.CreateOrUpdate(ctx, app.Config.Identity.ResourceGroup, groupsParams)
	if err != nil {
		return err
	}

	app.Logf("Identity: creating user assigned identity: %s", app.Config.Identity.IdentityName)
	msiClient := msi.NewUserAssignedIdentitiesClient(app.Config.SubscriptionID)
	msiClient.Authorizer = authorizer
	identityParams := msi.Identity{
		Location: &app.Config.Identity.Location,
	}
	_, err = msiClient.CreateOrUpdate(ctx, app.Config.Identity.ResourceGroup, app.Config.Identity.IdentityName, identityParams)
	if err != nil {
		return err
	}

	return app.IdentityGet(ctx)
}
