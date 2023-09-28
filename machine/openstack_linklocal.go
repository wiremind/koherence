package machine

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	openstackDataService = "http://169.254.169.254/openstack/latest/"
	metadataJsonService  = openstackDataService + "meta_data.json"
)

/*
 * This struct will contain the data returned from metadataJsonService service.
 *
 * Example on an openstack machine:
 * ```
 * # curl -s http://169.254.169.254/openstack/latest/meta_data.json | jq
 *{
 *   "uuid": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
 *   "meta": {...},
 *   "hostname": "my_node_hostname",
 *   "name": "my_node_name",
 *   "launch_index": 0,
 *   "availability_zone": "nova",
 *   "random_seed": "...",
 *   "project_id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
 *   "devices": []
 *   }
 * ```
 */
type openstackMetadata struct {
	Uuid             uuid.UUID `json:"uuid"`
	Hostname         string    `json:"hostname"`
	Name             string    `json:"name"`
	LaunchIndex      int       `json:"launch_index"`
	AvailabilityZone string    `json:"availability_zone"`
	RandomSeed       string    `json:"random_seed"`
	ProjectId        string    `json:"project_id"`
}

func fetchMetadataJson() (*openstackMetadata, error) {
	var metadatas openstackMetadata

	// Build request
	req, err := http.NewRequest("GET", metadataJsonService, nil)
	if err != nil {
		slog.Error(
			"Could not build request",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// We only want json, if not there is an issue with Openstack
	req.Header.Set("Accept", "application/json")

	// Do request
	httpClient := &http.Client{Timeout: 5 * time.Second}

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error(
			"Error requesting Openstack metadata service.",
			slog.String("url", metadataJsonService),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&metadatas); err != nil {
		slog.Error(
			"Error when decoding Openstack metadata json.",
			//slog.String("json", string(b)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	slog.Debug(
		"Metadata fetched from Openstack",
		slog.Any("struct", metadatas),
	)

	return &metadatas, nil
}

func ReadOpenstackInfos() (*MachineInfos, error) {
	metadatas, err := fetchMetadataJson()

	if err != nil {
		panic(err)
	}

	return &MachineInfos{
		Uuid:     metadatas.Uuid,
		Hostname: metadatas.Hostname,
		// TODO just leave empty ?
		SysVendor:     "OpenStack Foundation",
		ChassisVendor: "QEMU",
	}, nil
}
