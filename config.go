package main

type StorageConfig struct {
	Location      string
	ResourceGroup string
	AccountName   string
	AccountID     string
	ContainerName string
	ContainerID   string
}

type IdentityConfig struct {
	Location      string
	ResourceGroup string
	IdentityName  string
	IdentityID    string
}

type MachineConfig struct {
	Location      string
	ResourceGroup string
	MachineName   string
	MachineID     string
	Files         []string
	Command       string
}

type ImageConfig struct {
	Location      string
	ResourceGroup string
	ImageName     string
	ImageID       string
}

type GalleryConfig struct {
	Location           string
	ResourceGroup      string
	GalleryName        string
	GalleryID          string
	GalleryImageName   string
	GalleryImageID     string
	Publisher          string
	Offer              string
	SKU                string
	OSState            string
	OSType             string
	ReplicationRegions []string
	ExcludeFromLatest  bool
	StorageAccountType string
}

type BuilderConfig struct {
	Location      string
	ResourceGroup string
	BuilderName   string
	BuilderID     string
}

type AppConfig struct {
	TenantID       string
	ClientID       string
	SubscriptionID string
	Storage        StorageConfig
	Identity       IdentityConfig
	Machine        MachineConfig
	Image          ImageConfig
	Gallery        GalleryConfig
	Builder        BuilderConfig
}
