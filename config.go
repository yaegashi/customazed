package main

// StorageConfig is configuration for storage account and blob container
type StorageConfig struct {
	Location      string `json:"location,omitempty"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
	AccountName   string `json:"accountName,omitempty"`
	AccountID     string `json:"accountId,omitempty"`
	ContainerName string `json:"containerName,omitempty"`
	ContainerID   string `json:"containerId,omitempty"`
	Prefix        string `json:"prefix,omitempty"`
}

// IdentityConfig is configuration for user assigned identity
type IdentityConfig struct {
	Location      string `json:"location,omitempty"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
	IdentityName  string `json:"identityName,omitempty"`
	IdentityID    string `json:"identityId,omitempty"`
}

// MachineConfig is configuration for virtual machine
type MachineConfig struct {
	ResourceGroup string `json:"resourceGroup,omitempty"`
	MachineName   string `json:"machineName,omitempty"`
	MachineID     string `json:"machineId,omitempty"`
}

// ImageConfig is configuration for managed image
type ImageConfig struct {
	Location      string `json:"location,omitempty"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
	ImageName     string `json:"imageName,omitempty"`
	ImageID       string `json:"imageId,omitempty"`
	SkipSetup     bool   `json:"skipSetup,omitempty"`
	SkipCreate    bool   `json:"skipCreate,omitempty"`
}

// GalleryConfig is configuration for shared image gallery
type GalleryConfig struct {
	Location           string   `json:"location,omitempty"`
	ResourceGroup      string   `json:"resourceGroup,omitempty"`
	GalleryName        string   `json:"galleryName,omitempty"`
	GalleryID          string   `json:"galleryId,omitempty"`
	GalleryImageName   string   `json:"galleryImageName,omitempty"`
	GalleryImageID     string   `json:"galleryImageId,omitempty"`
	Publisher          string   `json:"publisher,omitempty"`
	Offer              string   `json:"offer,omitempty"`
	SKU                string   `json:"sku,omitempty"`
	OSState            string   `json:"osState,omitempty"`
	OSType             string   `json:"osType,omitempty"`
	HyperVGeneration   string   `json:"hyperVGeneration,omitempty"`
	ReplicationRegions []string `json:"replicationRegions,omitempty"`
	ExcludeFromLatest  bool     `json:"excludeFromLatest,omitempty"`
	StorageAccountType string   `json:"storageAccountType,omitempty"`
	SkipSetup          bool     `json:"skipSetup,omitempty"`
	SkipCreate         bool     `json:"skipCreate,omitempty"`
}

// BuilderConfig is configuration for image template
type BuilderConfig struct {
	Location      string `json:"location,omitempty"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
	BuilderName   string `json:"builderName,omitempty"`
	BuilderID     string `json:"builderId,omitempty"`
}

// Config is configuration for application
type Config struct {
	ID             string            `json:"id,omitempty"`
	TenantID       string            `json:"tenantId,omitempty"`
	ClientID       string            `json:"clientId,omitempty"`
	SubscriptionID string            `json:"subscriptionId,omitempty"`
	HashNS         string            `json:"hashNS,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
	Storage        StorageConfig     `json:"storage,omitempty"`
	Identity       IdentityConfig    `json:"identity,omitempty"`
	Machine        MachineConfig     `json:"machine,omitempty"`
	Image          ImageConfig       `json:"image,omitempty"`
	Gallery        GalleryConfig     `json:"gallery,omitempty"`
	Builder        BuilderConfig     `json:"builder,omitempty"`
}
