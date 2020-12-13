package main

import (
	"context"

	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppSetup is app setup command
type AppSetup struct {
	*App
}

// AppSetupCmder returns Cmder for app setup
func (app *App) AppSetupCmder() cmder.Cmder {
	return &AppSetup{App: app}
}

// Cmd returns Command for app setup
func (app *AppSetup) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "setup",
		Short:        "Customazed setup",
		RunE:         app.RunE,
		SilenceUsage: true,
	}
	return cmd
}

// RunE is main routine for app setup
func (app *AppSetup) RunE(cmd *cobra.Command, args []string) error {
	var err error
	_, err = app.ARMToken()
	if err != nil {
		return err
	}
	_, err = app.StorageToken()
	if err != nil {
		return err
	}
	ctx := context.Background()
	err = app.StorageSetup(ctx)
	if err != nil {
		return err
	}
	err = app.IdentitySetup(ctx)
	if err != nil {
		return err
	}
	err = app.MachineSetup(ctx)
	if err != nil {
		return err
	}
	err = app.ImageSetup(ctx)
	if err != nil {
		return err
	}
	err = app.GallerySetup(ctx)
	if err != nil {
		return err
	}
	err = app.BuilderSetup(ctx)
	if err != nil {
		return err
	}
	err = app.RoleSetup(ctx)
	if err != nil {
		return err
	}
	return nil
}
