//go:build linux
// +build linux

package machine

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

const (
	hostnamePath = "/etc/hostname"
	dmiInfosDir  = "/sys/devices/virtual/dmi/id/"
	dmiUuid      = dmiInfosDir + "product_uuid"
	/*
	 * For Openstack should return:
	 *
	 * ```bash
	 * # cat /sys/devices/virtual/dmi/id/sys_vendor
	 * OpenStack Foundation
	 * ```
	 */
	dmiSysVendor = dmiInfosDir + "sys_vendor"
	/*
	 * For VMs:
	 *
	 * ```bash
	 * # cat /sys/devices/virtual/dmi/id/chassis_vendor
	 * QEMU
	 ```
	*/
	dmiChassisVendor = dmiInfosDir + "chassis_vendor"
)

func readFsPath(path string) ([]byte, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		slog.Error(
			"Error reading file",
			slog.String("path", path),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return bytes.TrimSpace(buf), nil
}

// TODO improves this version, maybe check refcnt too ?
func getBSType() string {
	ret := "None"

	if _, err := os.Stat("/sys/module/sd_mod"); !os.IsNotExist(err) {
		ret = BsSCSI
	}

	if _, err := os.Stat("/sys/module/virtio_blk"); !os.IsNotExist(err) {
		ret = BsVirtioBlk
	}

	return ret
}

// ReadFsInfos reads machine infos directly from its pseudo-fs and return it in
// MachineInfos struct.
func ReadFsInfos() (*MachineInfos, error) {
	var hostname []byte
	var uuidBytes []byte
	var uuidT uuid.UUID
	var sysVendor []byte
	var chassisVendor []byte
	var err error

	if hostname, err = readFsPath(hostnamePath); err != nil {
		return nil, err
	}

	if uuidBytes, err = readFsPath(dmiUuid); err != nil {
		return nil, err
	}

	if uuidT, err = uuid.ParseBytes(uuidBytes); err != nil {
		slog.Error(
			"Error parsing machine uuid.",
			slog.String("uuid", string(uuidBytes)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if sysVendor, err = readFsPath(dmiSysVendor); err != nil {
		return nil, err
	}

	if chassisVendor, err = readFsPath(dmiChassisVendor); err != nil {
		return nil, err
	}

	return &MachineInfos{
		Uuid:             uuidT,
		Hostname:         string(hostname),
		SysVendor:        string(sysVendor),
		ChassisVendor:    string(chassisVendor),
		BlockStorageType: getBSType(),
	}, nil
}
