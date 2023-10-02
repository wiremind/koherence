//go:build linux
// +build linux

package bs

import (
	"bufio"
	"bytes"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/wiremind/koherence/machine"
)

const (
	udevData = "/run/udev/data/"
)

func commonDeviceSize(pseudoFsDevPath string) int {
	var ret int
	var err error

	buf, err := os.ReadFile(pseudoFsDevPath + "/size")
	if err != nil {
		slog.Error(
			"Could not get device size.",
			slog.String("error", err.Error()),
		)
		return 0
	}

	if ret, err = strconv.Atoi(string(bytes.TrimSpace(buf))); err != nil {
		slog.Error(
			"Could not convert string to int.",
			slog.String("number", string(buf)),
			slog.String("error", err.Error()),
		)
		return 0
	}

	return ret
}

func commonDeviceBlockDev(pseudoFsDevPath string) ([]byte, error) {
	buf, err := os.ReadFile(pseudoFsDevPath + "/dev")
	if err != nil {
		slog.Error(
			"Could not get device block dev.",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return bytes.TrimSpace(buf), nil
}

func virtioBlkDeviceCheck(pseudoFsDevPath string) bool {
	// Check driver just in case
	file, err := os.Open(pseudoFsDevPath + "/device/uevent")
	if err != nil {
		slog.Error(
			"Could not check device driver.",
			slog.String("error", err.Error()),
		)
		return false
	}
	defer file.Close()

	devtype := "None"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()

		if strings.HasPrefix(s, "DRIVER=") {
			split := strings.Split(s, "=")
			devtype = split[1]
			break
		}
	}

	if devtype != "virtio_blk" {
		slog.Error(
			"Wrong type of device found !",
			slog.String("device", pseudoFsDevPath),
			slog.String("devtype", devtype),
		)
		return false
	}

	return true
}

func virtioBlkDeviceUuid(blockDev string) (uuid.UUID, error) {
	var devUuid uuid.UUID

	path := udevData + "b" + blockDev

	file, err := os.Open(path)
	if err != nil {
		slog.Error(
			"Could not get device udev data.",
			slog.String("error", err.Error()),
		)
		panic("virtioBlkDeviceUuid")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()

		if strings.HasPrefix(s, "E:ID_SERIAL=") {
			split := strings.Split(s, "=")
			// XXX virtio-blk driver truncate serial (20bits instead of 36), we
			// just add 0 for now... We will get the full uuid via the provider
			// (partial-uuid check + size)
			if devUuid, err = uuid.Parse(split[1] + "000-000000000000"); err != nil {
				slog.Error(
					"Could not parse uuid.",
					slog.String("uuid", split[1]),
					slog.String("error", err.Error()),
				)
				return devUuid, err
			}
			break
		}
	}

	return devUuid, nil
}

func VirtioBlkDeviceInfos(device string) (*BlockStorageInfos, error) {
	var infos BlockStorageInfos
	var err error

	pseudoFsDevPath := "/sys/block/" + device

	// Check if device OK
	if _, err = os.Stat(pseudoFsDevPath); os.IsNotExist(err) {
		slog.Error(
			"Device does not exist.",
			slog.String("path", pseudoFsDevPath),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if !virtioBlkDeviceCheck(pseudoFsDevPath) {
		// TODO less panicky
		panic("VirtioBlkDeviceInfos")
	}
	infos.Device = device
	infos.Type = machine.BsVirtioBlk
	infos.Size = uint(commonDeviceSize(pseudoFsDevPath))

	blockDev, err := commonDeviceBlockDev(pseudoFsDevPath)
	if err != nil {
		// TODO less panicky
		panic("VirtioBlkDeviceInfos")
	}
	infos.blockDev = string(blockDev)

	devUuid, err := virtioBlkDeviceUuid(infos.blockDev)
	if err != nil {
		// TODO less panicky
		panic("VirtioBlkDeviceInfos")
	}
	infos.Uuid = devUuid

	return &infos, nil
}

func scsiDeviceCheck(pseudoFsDevPath string) bool {
	// Check driver just in case
	file, err := os.Open(pseudoFsDevPath + "/device/uevent")
	if err != nil {
		slog.Error(
			"Could not check device driver.",
			slog.String("error", err.Error()),
		)
		return false
	}
	defer file.Close()

	devtype := "None"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()

		if strings.HasPrefix(s, "DEVTYPE=") {
			split := strings.Split(s, "=")
			devtype = split[1]
			break
		}
	}

	if devtype != "scsi_device" {
		slog.Error(
			"Wrong type of device found !",
			slog.String("device", pseudoFsDevPath),
			slog.String("devtype", devtype),
		)
		return false
	}

	return true
}

func scsiDeviceUuid(blockDev string) (uuid.UUID, error) {
	var devUuid uuid.UUID

	path := udevData + "b" + blockDev

	file, err := os.Open(path)
	if err != nil {
		slog.Error(
			"Could not get device udev data.",
			slog.String("error", err.Error()),
		)
		panic("scsiDeviceUuid")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()

		if strings.HasPrefix(s, "E:ID_SCSI_SERIAL=") {
			split := strings.Split(s, "=")
			if devUuid, err = uuid.Parse(split[1]); err != nil {
				slog.Error(
					"Could not parse uuid.",
					slog.String("uuid", split[1]),
					slog.String("error", err.Error()),
				)
				return devUuid, err
			}
			break
		}
	}

	return devUuid, nil
}

func ScsiDeviceInfos(device string) (*BlockStorageInfos, error) {
	var infos BlockStorageInfos
	var err error

	pseudoFsDevPath := "/sys/block/" + device

	// Check if device OK
	if _, err = os.Stat(pseudoFsDevPath); os.IsNotExist(err) {
		slog.Error(
			"Device does not exist.",
			slog.String("path", pseudoFsDevPath),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if !scsiDeviceCheck(pseudoFsDevPath) {
		// TODO less panicky
		panic("ScsiDeviceInfos")
	}
	infos.Device = device
	infos.Type = machine.BsSCSI
	infos.Size = uint(commonDeviceSize(pseudoFsDevPath))

	blockDev, err := commonDeviceBlockDev(pseudoFsDevPath)
	if err != nil {
		// TODO less panicky
		panic("ScsiDeviceInfos")
	}
	infos.blockDev = string(blockDev)

	devUuid, err := scsiDeviceUuid(infos.blockDev)
	if err != nil {
		// TODO less panicky
		panic("ScsiDeviceInfos")
	}
	infos.Uuid = devUuid
	infos.FullUuid = devUuid

	return &infos, nil
}

//func extractBsInfos(machineInfos *machine.MachineInfos) (*BlockStorageInfos, error) {
//	var infos *BlockStorageInfos
//	var err error
//
//	switch machineInfos.BlockStorageType {
//	case machine.BsSCSI:
//		infos, err = extractSCSIInfos()
//	case machine.BsVirtioBlk:
//		infos, err = extractVirtioBlkInfos()
//	default:
//		slog.Error(
//			"Do not support this block-storage type.",
//			slog.String("type", machineInfos.BlockStorageType),
//		)
//		// TODO do better
//		panic("extractBsInfos")
//	}
//	return infos, nil
//}
