package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/features"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppFeatureShow is app feature show command
type AppFeatureShow struct {
	*AppFeature
}

// AppFeatureShowCmder returns Cmder for app feature show
func (app *AppFeature) AppFeatureShowCmder() cmder.Cmder {
	return &AppFeatureShow{AppFeature: app}
}

// Cmd returns Command for app feature show
func (app *AppFeatureShow) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show",
		Short:        "Show Azure features/providers",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app feature show
func (app *AppFeatureShow) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	featuresClient := features.NewClient(app.Config.SubscriptionID)
	featuresClient.Authorizer = authorizer
	for _, f := range appFeatureFeatures {
		feature, err := featuresClient.Get(ctx, f[0], f[1])
		if err != nil {
			return err
		}
		app.Logf("Feature: %s/%s: %s", f[0], f[1], *feature.Properties.State)
	}

	providersClient := resources.NewProvidersClient(app.Config.SubscriptionID)
	providersClient.Authorizer = authorizer
	for _, p := range appFeatureProviders {
		provider, err := providersClient.Get(ctx, p, "")
		if err != nil {
			return err
		}
		app.Logf("Provider: %s: %s", p, *provider.RegistrationState)
	}

	return nil
}
