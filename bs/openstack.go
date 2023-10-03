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

// XXX we only use environment variables method
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

func OpenstackGetBlockStorage(machineUuid uuid.UUID) (map[uuid.UUID]*BlockStorageInfos, error) {
	var openstackMachineUuid uuid.UUID
	var device string
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

	blockStorages := map[uuid.UUID]*BlockStorageInfos{}

	// TODO goroutine ?
	for _, vol := range allVolumes {
		if vol.Multiattach || len(vol.Attachments) > 1 {
			slog.Error(
				"Detected multi-attach for a volume.",
				slog.String("volumeID", vol.ID),
				slog.Any("attachments", vol.Attachments),
			)
			// TODO do better
			panic("multiattach")
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

		blockStorages[volUuid] = &BlockStorageInfos{
			Uuid:      volUuid,
			FullUuid:  volUuid,
			MachineId: openstackMachineUuid,
			Size:      uint(vol.Size),
			Device:    device,
			Type:      machine.BsUnknown,
			blockDev:  "unknown",
			Status:    vol.Status,
			Metadata:  vol.Metadata,
		}
	}

	return blockStorages, nil
}
