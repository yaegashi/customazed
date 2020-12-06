package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

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
	cmd.Flags().StringVarP(&app.Input, "input", "i", "-", "input file path")
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

	b, err := ioutil.ReadFile(app.Input)
	if err != nil {
		return err
	}

	var template virtualmachineimagebuilder.ImageTemplate
	err = json.Unmarshal(b, &template)
	if err != nil {
		return err
	}

	template.Location = &app.Config.Builder.Location
	{
		// Workaround for validation failure
		// Similar issue: https://github.com/Azure/azure-sdk-for-go/issues/2445
		if s, ok := template.Source.AsImageTemplateManagedImageSource(); ok {
			template.Source = s
		}
		if s, ok := template.Source.AsImageTemplatePlatformImageSource(); ok {
			template.Source = s
		}
		if s, ok := template.Source.AsImageTemplateSharedImageVersionSource(); ok {
			template.Source = s
		}
	}
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
	su := app.NewStorageUploader(ctx)
	var customizes []virtualmachineimagebuilder.BasicImageTemplateCustomizer
	for _, customize := range *template.Customize {
		if c, ok := customize.AsImageTemplateShellCustomizer(); ok {
			if c.ScriptURI != nil {
				u, err := su.Add(*c.ScriptURI)
				if err != nil {
					return err
				}
				c.ScriptURI = &u
			}
			customizes = append(customizes, c)
		} else if c, ok := customize.AsImageTemplatePowerShellCustomizer(); ok {
			if c.ScriptURI != nil {
				u, err := su.Add(*c.ScriptURI)
				if err != nil {
					return err
				}
				c.ScriptURI = &u
			}
			customizes = append(customizes, c)
		} else if c, ok := customize.AsImageTemplateFileCustomizer(); ok {
			if c.SourceURI != nil {
				u, err := su.Add(*c.SourceURI)
				if err != nil {
					return err
				}
				c.SourceURI = &u
			}
			customizes = append(customizes, c)
		}
	}
	template.Customize = &customizes
	app.Dump(template)
	su.Execute(ctx)

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
