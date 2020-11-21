package main

import (
	"context"

	"github.com/yaegashi/customazed/utils"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/msi/mgmt/msi"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
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

func (app *App) IdentityGet(ctx context.Context) error {
	cfgIdentity := app.Config.Identity
	if utils.HasEmpty(cfgIdentity.Location, cfgIdentity.ResourceGroup, cfgIdentity.IdentityName) {
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
	cfgIdentity := app.Config.Identity
	if utils.HasEmpty(cfgIdentity.Location, cfgIdentity.ResourceGroup, cfgIdentity.IdentityName) {
		app.Log("Identity: missing configuration")
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
