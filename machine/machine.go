package machine

import (
	"github.com/google/uuid"
)

type MachineInfos struct {
	Uuid          uuid.UUID
	Hostname      string
	SysVendor     string
	ChassisVendor string
}
