package bs

import (
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/wiremind/koherence/machine"
)

// XXX we only use environment variables method, needs:
// OS_AUTH_URL, OS_DOMAIN_ID or OS_DOMAIN_NAME, OS_PROJECT_ID, OS_PROJECT_NAME,
// OS_USERNAME, OS_PASSWORD, OS_REGION_NAME, OS_INTERFACE,
// OS_IDENTITY_API_VERSION

func openstackVolumesServiceAuth() (*gophercloud.ServiceClient, error) {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		slog.Error(
			"Error during openstack.AuthOptionsFromEnv().",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		slog.Error(
			"Error during openstack.AuthenticatedClient().",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	client, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
	if err != nil {
		slog.Error(
			"Error during openstack.NewBlockStorageV3().",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return client, nil
}

type MultiAttachments struct {
	MultiAttach MultiAttach `json:"multi-attach"`
}

type MultiAttach struct {
	Count int               `json:"count"`
	Data  []OpenstackAttach `json:"data"`
}

type OpenstackAttach struct {
	AttachmentId string `json:"attachment_id"`
	Device       string `json:"device"`
	HostName     string `json:"host_name"`
	ServerId     string `json:"server_id"`
	VolumeId     string `json:"volume_id"`
}

func openstackAllVolumes() ([]volumes.Volume, error) {
	var err error

	client, err := openstackVolumesServiceAuth()
	if err != nil {
		slog.Error(
			"Error during openstackServiceAuth(\"volumes\").",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// XXX we check all volumes each time without filter cause we do not trust
	// Openstack databases. DO NOT FILTER BY SERVER
	allPages, err := volumes.List(client, volumes.ListOpts{}).AllPages()
	if err != nil {
		slog.Error(
			"Error during volumes.List().",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	allVolumes, err := volumes.ExtractVolumes(allPages)
	if err != nil {
		slog.Error(
			"Error during volumes.ExtractVolumes().",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return allVolumes, nil
}

func OpenstackGetMultiAttachments() (*MultiAttachments, error) {
	var err error

	allVolumes, err := openstackAllVolumes()
	if err != nil {
		return nil, err
	}

	attachments := []OpenstackAttach{}

	for _, vol := range allVolumes {
		if vol.Multiattach || len(vol.Attachments) > 1 {
			for _, attachment := range vol.Attachments {
				openstack_attach := OpenstackAttach{
					AttachmentId: attachment.AttachmentID,
					Device:       attachment.Device,
					HostName:     attachment.HostName,
					ServerId:     attachment.ServerID,
					VolumeId:     attachment.VolumeID,
				}
				attachments = append(attachments, openstack_attach)
			}
		}
	}

	multiAttachments := MultiAttachments{
		MultiAttach: MultiAttach{
			Count: len(attachments),
			Data:  attachments,
		},
	}

	return &multiAttachments, nil
}

func OpenstackGetBlockStorage(machineInfos *machine.MachineInfos) (map[uuid.UUID]*BlockStorageInfos, error) {
	var openstackMachineUuid uuid.UUID
	var device string
	var err error

	allVolumes, err := openstackAllVolumes()
	if err != nil {
		return nil, err
	}

	blockStorages := map[uuid.UUID]*BlockStorageInfos{}

	// TODO goroutine ?
	for _, vol := range allVolumes {
		if vol.Multiattach || len(vol.Attachments) > 1 {
			slog.Warn(
				"Detected multi-attach for a volume.",
				slog.String("volumeID", vol.ID),
				slog.Any("attachments", vol.Attachments),
			)
			// TODO do better
			continue
		}

		volUuid, err := uuid.Parse(vol.ID)
		if err != nil {
			slog.Error(
				"Could not parse volume UUID.",
				slog.String("uuid", vol.ID),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
		realUuid := volUuid

		if len(vol.Attachments) == 0 {
			openstackMachineUuid = uuid.Nil
			device = "nil"
		} else {
			device = vol.Attachments[0].Device

			openstackMachineUuid, err = uuid.Parse(vol.Attachments[0].ServerID)
			if err != nil {
				slog.Error(
					"Could not parse machine UUID.",
					slog.String("uuid", vol.Attachments[0].ServerID),
					slog.String("error", err.Error()),
				)
				return nil, err
			}
		}

		// Virtio-blk driver truncates block-storage UUID, taking care of it.
		// We match on partial-UUID (20bytes + size).
		if machineInfos.BlockStorageType == machine.BsVirtioBlk {
			virtioBlkUuidMask := uuid.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

			for i := 8; i < 16; i++ {
				volUuid[i] &= virtioBlkUuidMask[i]
			}

			slog.Debug(
				"Openstack truncated uuid",
				slog.String("uuid", volUuid.String()),
			)
		}

		blockStorages[volUuid] = &BlockStorageInfos{
			Uuid:      volUuid,
			FullUuid:  realUuid,
			MachineId: openstackMachineUuid,
			Size:      vol.Size, // Openstack returns directly in Gi
			Device:    device,
			Type:      machine.BsUnknown,
			BlockDev:  "unknown",
			Status:    vol.Status,
			Metadata:  vol.Metadata,
		}
	}

	return blockStorages, nil
}
