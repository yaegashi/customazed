# Customazed CLI - Azure VM Custom Script Helper

## Introduction

Customazed CLI makes it easy to run custom scripts on Azure VM along with various asset files.
The files are securely transfered from your local disk to VMs via blob storage proctected by RBAC and MSI infrastructure.

It supports the following Azure services:

- Azure VM Custom Script Extension
- Azure VM Image Builder
- ARM Templates (coming soon)

It automatically prepares the following resources and RBACs for you:

- Storage account with blob container
- Virtual machine
- Virtual machine image template
- User assigned identity
- Mangaed image
- Shared image gallery

## Usage

```text
Customazed CLI - Azure VM Custom Script Helper

Usage:
  customazed [command]

Available Commands:
  builder     Azure VM Image Builder
  config      Configuration
  help        Help about any command
  login       Force dev auth login
  machine     Azure VM Custom Script Extension
  setup       Customazed setup
  template    Customazed template

Flags:
      --auth string              auth source [dev,env,file] (env:CUSTOMAZED_AUTH, default:dev)
      --auth-dev string          auth dev store (env:CUSTOMAZED_AUTH_DEV, default:auth_dev.json)
      --auth-file string         auth file store (env:CUSTOMAZED_AUTH_FILE, default:auth_file.json)
      --client-id string         Azure client ID (env:AZURE_CLIENT_ID, default:a3c13aac-2eb7-4d8a-b7ae-c29b516d566b)
      --config-dir string        config dir (env:CUSTOMAZED_CONFIG_DIR, default:.customazed)
  -f, --config-file string       config file (env:CUSTOMAZED_CONFIG_FILE, default:customazed.json)
      --hash-ns string           Hash namespace (env:CUSTOMAZED_HASHNS, default:random)
  -h, --help                     help for customazed
      --no-login                 disable login
  -q, --quiet                    quiet
      --subscription-id string   Azure subscription ID (env:AZURE_SUBSCRIPTION_ID, default:)
      --tenant-id string         Azure tenant ID (env:AZURE_TENANT_ID, default:common)
  -v, --version                  version for customazed

Use "customazed [command] --help" for more information about a command.
```
