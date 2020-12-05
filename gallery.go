package main

import (
	"context"

	"github.com/yaegashi/customazed/utils/ssutil"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
)

func (app *App) Gallery(ctx context.Context) (*compute.Gallery, error) {
	if app._Gallery == nil {
		err := app.GalleryGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._Gallery, nil
}

func (app *App) GalleryImage(ctx context.Context) (*compute.GalleryImage, error) {
	if app._GalleryImage == nil {
		err := app.GalleryGet(ctx)
		if err != nil {
			return nil, err
		}
	}
	return app._GalleryImage, nil
}

func (app *App) GalleryValid() bool {
	cfg := app.Config.Gallery
	if ssutil.HasEmpty(cfg.Location, cfg.ResourceGroup, cfg.GalleryName, cfg.GalleryImageName, cfg.Publisher, cfg.Offer, cfg.SKU) {
		app.Log("Gallery: missing configuration")
		return false
	}
	return true
}

func (app *App) GalleryGet(ctx context.Context) error {
	if !app.GalleryValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	galleriesClient := compute.NewGalleriesClient(app.Config.SubscriptionID)
	galleriesClient.Authorizer = authorizer
	gallery, err := galleriesClient.Get(ctx, app.Config.Gallery.ResourceGroup, app.Config.Gallery.GalleryName)
	if err != nil {
		return err
	}

	galleryImagesClient := compute.NewGalleryImagesClient(app.Config.SubscriptionID)
	galleryImagesClient.Authorizer = authorizer
	galleryImage, err := galleryImagesClient.Get(ctx, app.Config.Gallery.ResourceGroup, app.Config.Gallery.GalleryName, app.Config.Gallery.GalleryImageName)
	if err != nil {
		return err
	}

	app._Gallery = &gallery
	app._GalleryImage = &galleryImage

	app.Config.Gallery.GalleryID = *gallery.ID
	app.Config.Gallery.GalleryImageID = *galleryImage.ID

	return nil
}

func (app *App) GallerySetup(ctx context.Context) error {
	if !app.GalleryValid() {
		return nil
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	app.Logf("Gallery: creating resource group: %s", app.Config.Gallery.ResourceGroup)
	groupsClient := resources.NewGroupsClient(app.Config.SubscriptionID)
	groupsClient.Authorizer = authorizer
	group := resources.Group{
		Location: &app.Config.Gallery.Location,
	}
	_, err = groupsClient.CreateOrUpdate(ctx, app.Config.Gallery.ResourceGroup, group)
	if err != nil {
		return err
	}

	app.Logf("Gallery: creating gallery: %s", app.Config.Gallery.GalleryName)
	galleriesClient := compute.NewGalleriesClient(app.Config.SubscriptionID)
	galleriesClient.Authorizer = authorizer
	gallery := compute.Gallery{
		Location: &app.Config.Gallery.Location,
	}
	galleryFuture, err := galleriesClient.CreateOrUpdate(ctx, app.Config.Gallery.ResourceGroup, app.Config.Gallery.GalleryName, gallery)
	if err != nil {
		return err
	}
	err = galleryFuture.WaitForCompletionRef(ctx, galleriesClient.Client)
	if err != nil {
		return err
	}

	app.Logf("Gallery: creating gallery image: %s", app.Config.Gallery.GalleryImageName)
	galleryImagesClient := compute.NewGalleryImagesClient(app.Config.SubscriptionID)
	galleryImagesClient.Authorizer = authorizer
	galleryImage := compute.GalleryImage{
		Location: &app.Config.Gallery.Location,
		GalleryImageProperties: &compute.GalleryImageProperties{
			Identifier: &compute.GalleryImageIdentifier{
				Publisher: &app.Config.Gallery.Publisher,
				Offer:     &app.Config.Gallery.Offer,
				Sku:       &app.Config.Gallery.SKU,
			},
			OsState: compute.OperatingSystemStateTypes(app.Config.Gallery.OSState),
			OsType:  compute.OperatingSystemTypes(app.Config.Gallery.OSType),
		},
	}
	galleryImageFuture, err := galleryImagesClient.CreateOrUpdate(ctx, app.Config.Gallery.ResourceGroup, app.Config.Gallery.GalleryName, app.Config.Gallery.GalleryImageName, galleryImage)
	if err != nil {
		return err
	}
	err = galleryImageFuture.WaitForCompletionRef(ctx, galleryImagesClient.Client)
	if err != nil {
		return err
	}

	err = app.GalleryGet(ctx)
	if err != nil {
		return err
	}

	return nil
}
