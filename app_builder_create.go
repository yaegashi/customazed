package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/virtualmachineimagebuilder/mgmt/virtualmachineimagebuilder"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

type AppBuilderCreate struct {
	*AppBuilder
	Input string
}

func (app *AppBuilder) AppBuilderCreateCmder() cmder.Cmder {
	return &AppBuilderCreate{AppBuilder: app}
}

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

	var b []byte
	if app.Input == "-" {
		b, err = ioutil.ReadAll(os.Stdin)
	} else {
		b, err = ioutil.ReadFile(app.Input)
	}
	if err != nil {
		return err
	}

	var template virtualmachineimagebuilder.ImageTemplate
	err = json.Unmarshal(b, &template)
	if err != nil {
		return err
	}

	template.Location = &app.Config.Builder.Location

	if identity != nil {
		template.Identity = &virtualmachineimagebuilder.ImageTemplateIdentity{
			Type: virtualmachineimagebuilder.UserAssigned,
			UserAssignedIdentities: map[string]*virtualmachineimagebuilder.ImageTemplateIdentityUserAssignedIdentitiesValue{
				*identity.ID: {},
			},
		}
	}

	var distributes []virtualmachineimagebuilder.BasicImageTemplateDistributor
	if image != nil {
		distributes = append(distributes, virtualmachineimagebuilder.ImageTemplateManagedImageDistributor{
			Type:          virtualmachineimagebuilder.TypeBasicImageTemplateDistributorTypeManagedImage,
			Location:      &app.Config.Image.Location,
			ImageID:       &app.Config.Image.ImageID,
			RunOutputName: to.StringPtr("ManagedImage"),
		})
	}
	if galleryImage != nil {
		distributes = append(distributes, virtualmachineimagebuilder.ImageTemplateSharedImageDistributor{
			GalleryImageID:     galleryImage.ID,
			ReplicationRegions: &app.Config.Gallery.ReplicationRegions,
			ExcludeFromLatest:  &app.Config.Gallery.ExcludeFromLatest,
			StorageAccountType: virtualmachineimagebuilder.SharedImageStorageAccountType(app.Config.Gallery.StorageAccountType),
			RunOutputName:      to.StringPtr("SharedImage"),
		})
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
	err = templateFuture.WaitForCompletionRef(ctx, templatesClient.Client)
	if err != nil {
		return err
	}

	return nil
}
