package run

import (
	"fmt"
	"log"
	"path/filepath"

	mkapi "github.com/hickeyma/helm-mapkubeapis/pkg/common"
	v2 "github.com/hickeyma/helm-mapkubeapis/pkg/v2"
	v3 "github.com/hickeyma/helm-mapkubeapis/pkg/v3"
	"github.com/mongodb-forks/drone-helm3/internal/env"
)

// MapKube holds the parameters to run the MapReleaseWithUnSupportedAPIs action
type MapKube struct {
	namespace   string
	kubeConfig  string
	kubeContext string
	RunV2       bool
	mapOptions  mkapi.MapOptions
}

// NewMapKube initializes MapKubeApi by using values from env.Config
func NewMapKube(cfg env.Config, kubeConfig, kubeContext string, runV2 bool) *MapKube {

	mapKube := &MapKube{
		namespace:   cfg.Namespace,
		kubeConfig:  kubeConfig,
		kubeContext: kubeContext,
		RunV2:       runV2,
	}

	mapKube.mapOptions = mkapi.MapOptions{
		DryRun:           cfg.DryRun,
		ReleaseName:      cfg.Release,
		ReleaseNamespace: cfg.Namespace,
		MapFile:          filepath.Join("assets", "mapconfig.yaml"),
		KubeConfig: mkapi.KubeConfig{
			File:    kubeConfig,
			Context: kubeContext,
		},
	}

	return mapKube
}

// Execute runs mapkubeapis
func (m *MapKube) Execute() error {

	log.Printf("Release '%s' will be checked for deprecated or removed Kubernetes APIs and will be updated if necessary to supported API versions.\n", m.mapOptions.ReleaseName)

	if m.RunV2 {
		if err := v2.MapReleaseWithUnSupportedAPIs(m.mapOptions); err != nil {
			return err
		}
	} else {
		if err := v3.MapReleaseWithUnSupportedAPIs(m.mapOptions); err != nil {
			return err
		}
	}

	log.Printf("Map of release '%s' deprecated or removed APIs to supported versions, completed successfully.\n", m.mapOptions.ReleaseName)

	return nil
}

// Prepare checks required inputs
func (m *MapKube) Prepare() error {

	if m.mapOptions.ReleaseName == "" {
		return fmt.Errorf("release is required")
	}

	if m.mapOptions.ReleaseNamespace == "" {
		return fmt.Errorf("namespace is required")
	}

	return nil
}
