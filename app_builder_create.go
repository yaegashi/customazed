package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/virtualmachineimagebuilder/mgmt/2020-02-14/virtualmachineimagebuilder"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"

	"github.com/yaegashi/customazed/utils/inpututil"
)

// AppBuilderCreate is app builder create command
type AppBuilderCreate struct {
	*AppBuilder
	Input string
}

// AppBuilderCreateCmder returns Cmder for app builder create
func (app *AppBuilder) AppBuilderCreateCmder() cmder.Cmder {
	return &AppBuilderCreate{AppBuilder: app}
}

// Cmd returns Command for app builder create
func (app *AppBuilderCreate) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create image template",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&app.Input, "input", "i", "customazed_builder.json", "input file path")
	return cmd
}

// RunE is main routine for app builder create
func (app *AppBuilderCreate) RunE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	identity, err := app.Identity(ctx)
	if err != nil {
		return err
	}

	image, err := app.Image(ctx)
	if err != nil {
		return err
	}

	galleryImage, err := app.GalleryImage(ctx)
	if err != nil {
		return err
	}

	var template virtualmachineimagebuilder.ImageTemplate
	err = inpututil.UnmarshalJSONC(app.Input, &template)
	if err != nil {
		return err
	}

	template.Location = &app.Config.Builder.Location

	if identity != nil {
		template.Identity = &virtualmachineimagebuilder.ImageTemplateIdentity{
			Type: virtualmachineimagebuilder.ResourceIdentityTypeUserAssigned,
			UserAssignedIdentities: map[string]*virtualmachineimagebuilder.ImageTemplateIdentityUserAssignedIdentitiesValue{
				*identity.ID: {},
			},
		}
	}

	var distributes []virtualmachineimagebuilder.BasicImageTemplateDistributor
	if image != nil && !app.Config.Image.SkipCreate {
		distributes = append(distributes, virtualmachineimagebuilder.ImageTemplateManagedImageDistributor{
			Type:          virtualmachineimagebuilder.TypeBasicImageTemplateDistributorTypeManagedImage,
			Location:      &app.Config.Image.Location,
			ImageID:       &app.Config.Image.ImageID,
			RunOutputName: to.StringPtr("ManagedImage"),
		})
	}
	if galleryImage != nil && !app.Config.Gallery.SkipCreate {
		distributes = append(distributes, virtualmachineimagebuilder.ImageTemplateSharedImageDistributor{
			GalleryImageID:     galleryImage.ID,
			ReplicationRegions: &app.Config.Gallery.ReplicationRegions,
			ExcludeFromLatest:  &app.Config.Gallery.ExcludeFromLatest,
			StorageAccountType: virtualmachineimagebuilder.SharedImageStorageAccountType(app.Config.Gallery.StorageAccountType),
			RunOutputName:      to.StringPtr("SharedImage"),
		})
	}
	if len(distributes) == 0 {
		return fmt.Errorf("no distribution to create")
	}
	template.Distribute = &distributes

	// Workaround for validation failure
	// Similar issue: https://github.com/Azure/azure-sdk-for-go/issues/2445
	if s, ok := template.Source.AsImageTemplateManagedImageSource(); ok {
		template.Source = s
	} else if s, ok := template.Source.AsImageTemplatePlatformImageSource(); ok {
		template.Source = s
	} else if s, ok := template.Source.AsImageTemplateSharedImageVersionSource(); ok {
		template.Source = s
	}

	// Workaround for validation failure
	var customizes []virtualmachineimagebuilder.BasicImageTemplateCustomizer
	for _, customize := range *template.Customize {
		if c, ok := customize.AsImageTemplateShellCustomizer(); ok {
			customizes = append(customizes, c)
		} else if c, ok := customize.AsImageTemplatePowerShellCustomizer(); ok {
			customizes = append(customizes, c)
		} else if c, ok := customize.AsImageTemplateFileCustomizer(); ok {
			customizes = append(customizes, c)
		} else {
			customizes = append(customizes, customize)
		}
	}
	template.Customize = &customizes

	su := app.NewStorageUploader(ctx)
	tv := app.NewTemplateVariable(su)
	err = tv.Resolve(&template)
	if err != nil {
		return err
	}

	app.Dump(template)
	app.LogBuilderName()
	app.Prompt("Files to upload: %d", su.Files())

	if su.Valid() && su.Files() > 0 {
		err = su.Execute(ctx)
		if err != nil {
			return err
		}
	}

	app.Log("Creating image template...")
	templatesClient := virtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(app.Config.SubscriptionID)
	templatesClient.Authorizer = authorizer
	templateFuture, err := templatesClient.CreateOrUpdate(ctx, template, app.Config.Builder.ResourceGroup, app.Config.Builder.BuilderName)
	if err != nil {
		return err
	}

	return app.WaitForCompletion(ctx, &templateFuture, templatesClient.Client)
}
