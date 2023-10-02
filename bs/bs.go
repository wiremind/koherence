package bs

import (
	"github.com/google/uuid"
)

type BlockStorageInfos struct {
	Uuid uuid.UUID
	// Some BS types truncate UUIDs (virtio_blk for example...)
	FullUuid uuid.UUID
	Size     uint
	Device   string
	Type     string
	blockDev string
}
