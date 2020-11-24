package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/yaegashi/customazed/store"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/msi/mgmt/msi"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/virtualmachineimagebuilder/mgmt/virtualmachineimagebuilder"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/spf13/cobra"
)

const (
	defaultClientID   = "a3c13aac-2eb7-4d8a-b7ae-c29b516d566b"
	defaultTenantID   = "common"
	environConfigFile = "CUSTOMAZED_CONFIG_FILE"
	defaultConfigFile = "customazed.json"
	environConfigDir  = "CUSTOMAZED_CONFIG_DIR"
	defaultConfigDir  = ".customazed"
	environAuth       = "CUSTOMAZED_AUTH"
	defaultAuth       = "dev"
	environAuthFile   = "CUSTOMAZED_AUTH_FILE"
	defaultAuthFile   = "auth_file.json"
	environAuthDev    = "CUSTOMAZED_AUTH_DEV"
	defaultAuthDev    = "auth_dev.json"
)

type App struct {
	Config         AppConfig
	ConfigStore    *store.Store
	ConfigFile     string
	ConfigDir      string
	TenantID       string
	ClientID       string
	SubscriptionID string
	Auth           string
	AuthFile       string
	AuthDev        string
	Quiet          bool

	_ARMToken         *adal.ServicePrincipalToken
	_StorageToken     *adal.ServicePrincipalToken
	_StorageAccount   *storage.Account
	_StorageContainer *storage.BlobContainer
	_StorageMap       map[string]string
	_Identity         *msi.Identity
	_Machine          *compute.VirtualMachine
	_Builder          *virtualmachineimagebuilder.ImageTemplate
	_Image            *compute.Image
	_Gallery          *compute.Gallery
	_GalleryImage     *compute.GalleryImage
}

func (app *App) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "customazed",
		Short:             "Customazed CLI - Azure VM Custom Script Helper",
		PersistentPreRunE: app.PersistentPreRunE,
		SilenceUsage:      true,
	}
	cmd.PersistentFlags().StringVarP(&app.ConfigFile, "config-file", "f", "", envHelp("config file", environConfigFile, defaultConfigFile))
	cmd.PersistentFlags().StringVarP(&app.ConfigDir, "config-dir", "", "", envHelp("config dir", environConfigDir, defaultConfigDir))
	cmd.PersistentFlags().StringVarP(&app.TenantID, "tenant-id", "", "", envHelp("Azure tenant ID", auth.TenantID, defaultTenantID))
	cmd.PersistentFlags().StringVarP(&app.ClientID, "client-id", "", "", envHelp("Azure client ID", auth.ClientID, defaultClientID))
	cmd.PersistentFlags().StringVarP(&app.SubscriptionID, "subscription-id", "", "", envHelp("Azure subscription ID", auth.SubscriptionID, ""))
	cmd.PersistentFlags().StringVarP(&app.Auth, "auth", "", "", envHelp("auth source [dev,env,file]", environAuth, defaultAuth))
	cmd.PersistentFlags().StringVarP(&app.AuthFile, "auth-file", "", "", envHelp("auth file store", environAuthFile, defaultAuthFile))
	cmd.PersistentFlags().StringVarP(&app.AuthDev, "auth-dev", "", "", envHelp("auth dev store", environAuthDev, defaultAuthDev))
	cmd.PersistentFlags().BoolVarP(&app.Quiet, "quiet", "q", false, "quiet")
	return cmd
}

func envDefault(val, env, def string) string {
	if val == "" {
		val = os.Getenv(env)
	}
	if val == "" {
		val = def
	}
	return val
}

func envHelp(msg, env, def string) string {
	return fmt.Sprintf(`%s (env:%s, default:%s)`, msg, env, def)
}

func (app *App) PersistentPreRunE(cmd *cobra.Command, args []string) error {
	app.TenantID = envDefault(app.TenantID, auth.TenantID, defaultTenantID)
	app.ClientID = envDefault(app.ClientID, auth.ClientID, defaultClientID)
	app.SubscriptionID = envDefault(app.SubscriptionID, auth.SubscriptionID, "")
	app.ConfigFile = envDefault(app.ConfigFile, environConfigFile, defaultConfigFile)
	app.ConfigDir = envDefault(app.ConfigDir, environConfigDir, defaultConfigDir)
	app.Auth = envDefault(app.Auth, environAuth, defaultAuth)
	app.AuthDev = envDefault(app.AuthDev, environAuthDev, defaultAuthDev)
	app.AuthFile = envDefault(app.AuthFile, environAuthFile, defaultAuthFile)

	store, err := store.NewStore(app.ConfigDir)
	if err != nil {
		return err
	}
	app.ConfigStore = store

	app.Logf("Reading config file %s", app.ConfigFile)
	b, err := ioutil.ReadFile(app.ConfigFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &app.Config)
	if err != nil {
		return err
	}

	if app.Config.TenantID == "" {
		app.Config.TenantID = app.TenantID
	}
	if app.Config.ClientID == "" {
		app.Config.ClientID = app.ClientID
	}
	if app.Config.SubscriptionID == "" {
		app.Config.SubscriptionID = app.SubscriptionID
	}

	return nil
}

func (app *App) ARMAuthorizer() (autorest.Authorizer, error) {
	token, err := app.ARMToken()
	if err != nil {
		return nil, err
	}
	return autorest.NewBearerAuthorizer(token), nil
}

func (app *App) ARMToken() (*adal.ServicePrincipalToken, error) {
	if app._ARMToken == nil {
		token, err := app.GetToken()
		if err != nil {
			return nil, err
		}
		app._ARMToken = token
	}
	return app._ARMToken, nil
}

func (app *App) StorageToken() (*adal.ServicePrincipalToken, error) {
	if app._StorageToken == nil {
		token, err := app.GetTokenWithResource(azure.PublicCloud.ResourceIdentifiers.Storage)
		if err != nil {
			return nil, err
		}
		app._StorageToken = token
	}
	return app._StorageToken, nil
}

func (app *App) GetAuthorizer() (autorest.Authorizer, error) {
	return app.GetAuthorizerWithResource(azure.PublicCloud.ResourceManagerEndpoint)
}

func (app *App) GetAuthorizerWithResource(resource string) (autorest.Authorizer, error) {
	token, err := app.GetTokenWithResource(resource)
	if err != nil {
		return nil, err
	}
	return autorest.NewBearerAuthorizer(token), nil
}

func (app *App) GetToken() (*adal.ServicePrincipalToken, error) {
	return app.GetTokenWithResource(azure.PublicCloud.ResourceManagerEndpoint)
}

func (app *App) GetTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	token, err := app.AcquireToken()
	if err != nil {
		return nil, err
	}
	if token.Token().Resource != resource {
		err = token.RefreshExchange(resource)
		if err != nil {
			return nil, err
		}
	}
	return token, nil
}

func (app *App) AcquireToken() (*adal.ServicePrincipalToken, error) {
	switch app.Auth {
	case "env":
		settings, err := auth.GetSettingsFromEnvironment()
		if err != nil {
			return nil, err
		}
		app.Config.TenantID = settings.Values[auth.TenantID]
		app.Config.ClientID = settings.Values[auth.ClientID]
		if app.Config.SubscriptionID == "" {
			app.Config.SubscriptionID = settings.GetSubscriptionID()
		}
		if c, err := settings.GetClientCredentials(); err == nil {
			return c.ServicePrincipalToken()
		}
		if c, err := settings.GetClientCredentials(); err == nil {
			return c.ServicePrincipalToken()
		}
		if c, err := settings.GetUsernamePassword(); err == nil {
			return c.ServicePrincipalToken()
		}
		c := settings.GetMSI()
		app.Config.TenantID = "" // XXX: how to get tenant from MSI?
		app.Config.ClientID = c.ClientID
		return c.ServicePrincipalToken()
	case "file":
		loc, _ := app.ConfigStore.Location(app.AuthFile, true)
		app.Logf("Loading auth-file config in %s", loc)
		os.Setenv("AZURE_AUTH_LOCATION", loc)
		settings, err := auth.GetSettingsFromFile()
		if err != nil {
			return nil, err
		}
		app.Config.TenantID = settings.Values[auth.TenantID]
		app.Config.ClientID = settings.Values[auth.ClientID]
		if app.Config.SubscriptionID == "" {
			app.Config.SubscriptionID = settings.GetSubscriptionID()
		}
		if t, err := settings.ServicePrincipalTokenFromClientCredentials(azure.PublicCloud.ResourceManagerEndpoint); err == nil {
			return t, nil
		}
		if t, err := settings.ServicePrincipalTokenFromClientCertificate(azure.PublicCloud.ResourceManagerEndpoint); err == nil {
			return t, nil
		}
		return nil, errors.New("auth file missing client and certificate credentials")
	case "dev":
		loc, _ := app.ConfigStore.Location(app.AuthDev, true)
		app.Logf("Loading auth-dev token in %s", loc)
		b, err := app.ConfigStore.ReadFile(app.AuthDev)
		if err != nil {
			app.Logf("Warning: %s", err)
			return app.AuthorizeDeviceFlow()
		}
		var token *adal.ServicePrincipalToken
		err = json.Unmarshal(b, &token)
		if err != nil {
			app.Logf("Warning: %s", err)
			return app.AuthorizeDeviceFlow()
		}
		save := false
		token.SetRefreshCallbacks([]adal.TokenRefreshCallback{func(adal.Token) error { save = true; return nil }})
		err = token.EnsureFresh()
		if err != nil {
			app.Logf("Warning: %s", err)
			return app.AuthorizeDeviceFlow()
		}
		if save {
			b, err := json.Marshal(token)
			if err == nil {
				loc, _ := app.ConfigStore.Location(app.AuthDev, true)
				app.Logf("Saving auth-dev token in %s", loc)
				err = app.ConfigStore.WriteFile(app.AuthDev, b, 0600)
			}
			if err != nil {
				app.Logf("Warning: %s", err)
			}
		}
		return token, nil
	}
	return nil, fmt.Errorf("Unknown auth: %s", app.Auth)
}

func (app *App) AuthorizeDeviceFlow() (*adal.ServicePrincipalToken, error) {
	deviceConfig := auth.NewDeviceFlowConfig(app.Config.ClientID, app.Config.TenantID)
	token, err := deviceConfig.ServicePrincipalToken()
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(token)
	if err == nil {
		loc, _ := app.ConfigStore.Location(app.AuthDev, true)
		app.Logf("Saving auth-dev token in %s", loc)
		err = app.ConfigStore.WriteFile(app.AuthDev, b, 0600)
	}
	if err != nil {
		app.Logf("Warning: %s", err)
	}
	return token, nil
}

func (app *App) Log(args ...interface{}) {
	if !app.Quiet {
		log.Print(args...)
	}
}

func (app *App) Logln(args ...interface{}) {
	if !app.Quiet {
		log.Println(args...)
	}
}

func (app *App) Logf(format string, args ...interface{}) {
	if !app.Quiet {
		log.Printf(format, args...)
	}
}

func (app *App) Dump(v interface{}) {
	if !app.Quiet {
		b, err := json.MarshalIndent(v, "", "  ")
		if err == nil {
			log.Printf("\n%s", string(b))
		}
	}
}
