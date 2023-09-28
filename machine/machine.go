package machine

import (
	"github.com/google/uuid"
)

type MachineInfos struct {
	Uuid          uuid.UUID
	SysVendor     string
	ChassisVendor string
}
