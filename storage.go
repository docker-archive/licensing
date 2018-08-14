package licensing

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/licensing/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
)

var (
	licenseNamePrefix = "com.docker.license"
	licenseFilename   = "docker.lic"
)

// WrappedDockerClient provides methods useful for installing licenses to the wrapped docker engine or cluster
type WrappedDockerClient interface {
	NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error)
	ConfigCreate(ctx context.Context, config swarm.ConfigSpec) (types.ConfigCreateResponse, error)
	ConfigList(ctx context.Context, options types.ConfigListOptions) ([]swarm.Config, error)
}

// StoreLicense will store the license on the host filesystem and swarm (if swarm is active)
func StoreLicense(ctx context.Context, clnt WrappedDockerClient, license *model.IssuedLicense, rootDir string) error {

	licenseData, err := json.Marshal(*license)
	if err != nil {
		return err
	}

	// First determine if we're in swarm-mode or a stand-alone engine
	_, err = clnt.NodeList(ctx, types.NodeListOptions{})
	if err != nil { // TODO - check for the specific error message
		return writeLicenseToHost(ctx, clnt, licenseData, rootDir)
	}
	// Load this in the latest license index
	latestVersion, err := getLatestNamedConfig(clnt, licenseNamePrefix)
	if err != nil {
		return fmt.Errorf("unable to get latest license version: %s", err)
	}
	spec := swarm.ConfigSpec{
		Annotations: swarm.Annotations{
			Name: fmt.Sprintf("%s-%d", licenseNamePrefix, latestVersion+1),
			Labels: map[string]string{
				"com.docker.ucp.access.label":     "/",
				"com.docker.ucp.collection":       "swarm",
				"com.docker.ucp.collection.root":  "true",
				"com.docker.ucp.collection.swarm": "true",
			},
		},
		Data: licenseData,
	}
	_, err = clnt.ConfigCreate(context.Background(), spec)
	if err != nil {

		return fmt.Errorf("Failed to create license: %s", err)
	}

	return nil
}

// getLatestNamedConfig looks for versioned instances of configs with the
// given name prefix which have a `-NUM` integer version suffix. Returns the
// config with the higest version number found or nil if no such configs exist
// along with its version number.
func getLatestNamedConfig(dclient WrappedDockerClient, namePrefix string) (int, error) {
	latestVersion := -1
	// List any/all existing configs so that we create a newer version than
	// any that already exist.
	filter := filters.NewArgs()
	filter.Add("name", namePrefix)
	existingConfigs, err := dclient.ConfigList(context.Background(), types.ConfigListOptions{Filters: filter})
	if err != nil {
		return latestVersion, fmt.Errorf("unable to list existing configs: %s", err)
	}

	for _, existingConfig := range existingConfigs {
		existingConfigName := existingConfig.Spec.Name
		nameSuffix := strings.TrimPrefix(existingConfigName, namePrefix)
		if nameSuffix == "" || nameSuffix[0] != '-' {
			continue // No version specifier?
		}

		versionSuffix := nameSuffix[1:] // Trim the version separator.
		existingVersion, err := strconv.Atoi(versionSuffix)
		if err != nil {
			continue // Unable to parse version as integer.
		}
		if existingVersion > latestVersion {
			latestVersion = existingVersion
		}
	}

	return latestVersion, nil
}

func writeLicenseToHost(ctx context.Context, dclient WrappedDockerClient, license []byte, rootDir string) error {
	// TODO we should write the file out over the clnt instead of to the local filesystem
	return ioutil.WriteFile(filepath.Join(rootDir, LicenseFilename), license, 0644)
}
