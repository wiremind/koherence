//go:build linux
// +build linux

package bs

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/wiremind/koherence/machine"
)

const (
	udevData = "/run/udev/data/"
	sysBlock = "/sys/block/"
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

	/*
	 * XXX in linux size is stored in 512bytes sector, see:
	 * - https://unix.stackexchange.com/questions/512945/what-units-are-the-values-in-proc-partitions-and-sys-dev-block-block-size/512959#512959
	 * - https://unix.stackexchange.com/questions/555838/calculate-disk-byte-size-exclusively-from-sys-block
	 *
	 * We want to return in Gi unit so : * 512 / 1024 / 1024 / 1024 <=> >> 21
	 */
	return ret >> 21
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
			// XXX virtio-blk driver truncate serial (20bytes instead of 36), we
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

func virtioBlkDeviceInfos(device string) (*BlockStorageInfos, error) {
	var infos BlockStorageInfos
	var err error

	pseudoFsDevPath := sysBlock + device

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
	infos.Size = commonDeviceSize(pseudoFsDevPath)

	blockDev, err := commonDeviceBlockDev(pseudoFsDevPath)
	if err != nil {
		// TODO less panicky
		panic("VirtioBlkDeviceInfos")
	}
	infos.BlockDev = string(blockDev)

	devUuid, err := virtioBlkDeviceUuid(infos.BlockDev)
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

func scsiDeviceInfos(device string) (*BlockStorageInfos, error) {
	var infos BlockStorageInfos
	var err error

	pseudoFsDevPath := sysBlock + device

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
	infos.Size = commonDeviceSize(pseudoFsDevPath)

	blockDev, err := commonDeviceBlockDev(pseudoFsDevPath)
	if err != nil {
		// TODO less panicky
		panic("ScsiDeviceInfos")
	}
	infos.BlockDev = string(blockDev)

	devUuid, err := scsiDeviceUuid(infos.BlockDev)
	if err != nil {
		// TODO less panicky
		panic("ScsiDeviceInfos")
	}
	infos.Uuid = devUuid
	infos.FullUuid = devUuid

	return &infos, nil
}

func listDevices(bsType string) []string {
	var re *regexp.Regexp

	files := []string{}

	switch bsType {
	case machine.BsSCSI:
		re = regexp.MustCompile("sd[a-z]+")
	case machine.BsVirtioBlk:
		re = regexp.MustCompile("vd[a-z]")
	default:
		return files
	}

	walk := func(fn string, fi os.FileInfo, err error) error {
		if !re.MatchString(fn) {
			return nil
		}

		basename := filepath.Base(fn)
		files = append(files, basename)

		return nil
	}

	err := filepath.Walk(sysBlock, walk)
	if err != nil {
		// TODO do better
		panic(fmt.Sprintf("listDevices: %s", err))
	}

	return files
}

// TODO maybe get block-storage type directly without check instead of this
// double check machine/bs ?
func ExtractBsInfos(m *machine.MachineInfos) map[uuid.UUID]*BlockStorageInfos {
	var fn func(string) (*BlockStorageInfos, error)

	infos := map[uuid.UUID]*BlockStorageInfos{}
	devices := listDevices(m.BlockStorageType)

	switch m.BlockStorageType {
	case machine.BsSCSI:
		fn = scsiDeviceInfos
	case machine.BsVirtioBlk:
		fn = virtioBlkDeviceInfos
	default:
		// TODO do better
		panic("ExtractBsInfos")
	}

	for _, dev := range devices {
		i, err := fn(dev)
		if err != nil {
			// TODO do better
			panic(fmt.Sprintf("ExtractBsInfos: %s", err))
		}

		// Add machine UUID
		i.MachineId = m.Uuid

		// Add machine node name
		i.NodeName = m.KubeNodeName
		infos[i.Uuid] = i
	}

	return infos
}
