package machine

import (
	"github.com/google/uuid"
)

// Block Storage type enum
const (
	BsSCSI      string = "scsi"
	BsVirtioBlk        = "virtio_blk"
)

// Provider Enum
const (
	ProviderOpenstack string = "OpenStack Foundation"
)

type MachineInfos struct {
	Uuid             uuid.UUID
	Hostname         string
	SysVendor        string
	ChassisVendor    string
	BlockStorageType string
}
