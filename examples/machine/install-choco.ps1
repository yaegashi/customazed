$ErrorActionPreference = "Stop"

# Install chocolatey
Set-ExecutionPolicy Bypass -Scope Process -Force
Invoke-Expression ((New-Object System.Net.WebClient).DownloadString("https://chocolatey.org/install.ps1"))

# Install chocolatey packages
choco install -y --ignore-checksums googlechrome
choco install -y vscode
choco install -y git
choco install -y azure-cli
choco install -y az.powershell
choco install -y microsoftazurestorageexplorer
choco install -y nodejs.install
choco install -y rclone
