package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/spf13/cobra"
	cmder "github.com/yaegashi/cobra-cmder"
)

// AppBuilder is app builder command
type AppBuilder struct {
	*App
	Timeout string
}

// AppBuilderCmder returns Cmder for app builder
func (app *App) AppBuilderCmder() cmder.Cmder {
	return &AppBuilder{App: app}
}

// Cmd returns Command for app builder
func (app *AppBuilder) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "builder",
		Short:        "Azure VM Image Builder",
		SilenceUsage: true,
	}
	cmd.PersistentFlags().StringVarP(&app.Timeout, "timeout", "", "", "wait time out for completion")
	return cmd
}

// LogBuilderName shows current builder name
func (app *AppBuilder) LogBuilderName() {
	app.Logf("Current builder name: %s", app.Config.Builder.BuilderName)
}

// WaitForCompletion calls FutureAPI function with proper timeout setting
func (app *AppBuilder) WaitForCompletion(ctx context.Context, future azure.FutureAPI, client autorest.Client) error {
	if app.Timeout != "" {
		duration, err := time.ParseDuration(app.Timeout)
		if err != nil {
			return err
		}
		if duration == 0 {
			return nil
		}
		client.PollingDuration = duration
	}
	app.Logf("Waiting up to %s for completion...", client.PollingDuration)
	err := future.WaitForCompletionRef(ctx, client)
	if err != nil {
		aErr, ok := err.(autorest.DetailedError)
		if !ok || aErr.Original != context.DeadlineExceeded {
			return err
		}
		return fmt.Errorf("timed out")
	}
	app.Logf("Done")
	return nil
}
