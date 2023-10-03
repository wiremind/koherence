package machine

import (
	"github.com/google/uuid"
)

// Block Storage type enum
const (
	BsUnknown   string = "unknown"
	BsSCSI             = "scsi"
	BsVirtioBlk        = "virtio_blk"
)

// Provider Enum
const (
	ProviderOpenstack string = "OpenStack Foundation"
)

type MachineInfos struct {
	Uuid             uuid.UUID `json:"uuid"`
	Hostname         string    `json:"hostname"`
	SysVendor        string    `json:"sys_vendor"`
	ChassisVendor    string    `json:"chassis_vendor"`
	BlockStorageType string    `json:"block_storage_type"`
}
