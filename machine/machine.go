package machine

import (
	"github.com/google/uuid"
)

// Block Storage type enum
const (
	BsUnknown   string = "unknown"
	BsSCSI      string = "scsi"
	BsVirtioBlk string = "virtio_blk"
)

// Provider Enum
const (
	ProviderUnknown   string = "Unknown"
	ProviderOpenstack string = "OpenStack Foundation"
)

type MachineInfos struct {
	Uuid             uuid.UUID `json:"uuid"`
	Hostname         string    `json:"hostname"`
	SysVendor        string    `json:"sys_vendor"`
	ChassisVendor    string    `json:"chassis_vendor"`
	BlockStorageType string    `json:"block_storage_type"`
	KubeNodeName     string    `json:"node_name"`
}
