package main

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/features"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppFeatureRegister is app feature register command
type AppFeatureRegister struct {
	*AppFeature
}

// AppFeatureRegisterCmder returns Cmder for app feature register
func (app *AppFeature) AppFeatureRegisterCmder() cmder.Cmder {
	return &AppFeatureRegister{AppFeature: app}
}

// Cmd returns Command for app feature register
func (app *AppFeatureRegister) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "register",
		Short:        "Register Azure features/providers",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app feature register
func (app *AppFeatureRegister) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	featuresClient := features.NewClient(app.Config.SubscriptionID)
	featuresClient.Authorizer = authorizer
	for _, f := range appFeatureFeatures {
		app.Logf("Feature: registering %s/%s", f[0], f[1])
		_, err := featuresClient.Register(ctx, f[0], f[1])
		if err != nil {
			return err
		}
	}

	providersClient := resources.NewProvidersClient(app.Config.SubscriptionID)
	providersClient.Authorizer = authorizer
	for _, p := range appFeatureProviders {
		app.Logf("Provider: registering %s", p)
		_, err := providersClient.Register(ctx, p)
		if err != nil {
			return err
		}
	}

	return nil
}
