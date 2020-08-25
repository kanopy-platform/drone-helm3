package run

import (
	"errors"
	"fmt"
	"log"

	convertcmd "github.com/helm/helm-2to3/cmd"
	"github.com/helm/helm-2to3/pkg/common"
	"github.com/pelotech/drone-helm3/internal/env"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func v3ReleaseFound(release string, cfg *action.Configuration) bool {

	// check if release exists
	_, err := cfg.Releases.Deployed(release)
	switch {
	// case errors.Is(err, driver.ErrReleaseNotFound), errors.Is(err, driver.ErrNoDeployedReleases):
	case errors.Is(err, driver.ErrReleaseNotFound):
		log.Printf("No v3 Release of %s found", release)
	case err == nil:
		log.Printf("A v3 Release of %s was found", release)
		return true
	}

	return false
}

// Convert holds the parameters to run the Convert action
type Convert struct {
	namespace      string
	debug          action.DebugLog
	kubeConfig     string
	kubeContext    string
	convertOptions convertcmd.ConvertOptions
}

// NewConvert initialize Convert by using values from env.Config
func NewConvert(cfg env.Config, kubeConfig string, kubeContext string) *Convert {

	convert := &Convert{
		namespace:   cfg.Namespace,
		kubeConfig:  kubeConfig,
		kubeContext: kubeContext,
	}

	convert.convertOptions = convertcmd.ConvertOptions{
		DeleteRelease:      cfg.DeleteV2Releases,
		DryRun:             cfg.DryRun,
		MaxReleaseVersions: cfg.ReleaseVersionsMax,
		ReleaseName:        cfg.Release,
		StorageType:        "configmap",
		TillerLabel:        cfg.TillerLabel,
		TillerNamespace:    cfg.TillerNS,
		TillerOutCluster:   false,
	}

	if cfg.Debug {
		convert.debug = func(format string, v ...interface{}) {
			format = fmt.Sprintf("[debug] %s\n", format)
			_ = log.Output(2, fmt.Sprintf(format, v...))
		}
	}

	return convert
}

// Execute runs Convert from 2to3 package
// If a v2 version doesn't exists then convertcmd.Convert will error
// If a V3 version exists, we assume that was migrated and the conversion is not run
func (c *Convert) Execute() error {

	release := c.convertOptions.ReleaseName

	settings := cli.New()
	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(settings.RESTClientGetter(), c.namespace, "secrets", c.debug); err != nil {
		return err
	}

	// If there's already a v3 Release, migration shouldn't run
	if v3ReleaseFound(release, actionCfg) {
		return nil
	}

	kc := common.KubeConfig{
		File:    c.kubeConfig,
		Context: c.kubeContext,
	}

	return convertcmd.Convert(c.convertOptions, kc)
}

// Prepare is not used but it's required to fulfill the Step interface
func (c *Convert) Prepare() error {
	return nil
}
