package bs

import (
	"github.com/google/uuid"
)

type BlockStorageInfos struct {
	Uuid      uuid.UUID         `json:"uuid"`
	FullUuid  uuid.UUID         `json:"full_uuid"` // Some BS types truncate UUIDs (virtio_blk for example...)
	MachineId uuid.UUID         `json:"machine_id"`
	Size      int               `json:"size"` // Stored in Gi unit
	Device    string            `json:"device"`
	Type      string            `json:"type"`
	BlockDev  string            `json:"block_dev"`
	Status    string            `json:"status"`
	NodeName  string            `json:"node_name"`
	Metadata  map[string]string `json:"metadata"`
}
