package run

import (
	"fmt"
	"log"

	convertcmd "github.com/helm/helm-2to3/cmd"
	"github.com/helm/helm-2to3/pkg/common"
	"github.com/mongodb-forks/drone-helm3/internal/env"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func v3ReleaseFound(release string, cfg *action.Configuration) bool {

	if _, err := cfg.Releases.Deployed(release); err == nil {
		log.Printf("A v3 Release of %s was found", release)
		return true
	}

	log.Printf("No v3 Release of %s found", release)
	return false
}

// clientsetFromFile returns a ready-to-use client from a kubeconfig file
func clientsetFromFile(path string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load admin kubeconfig")
	}

	overrides := clientcmd.ConfigOverrides{Timeout: "15s"}
	clientConfig, err := clientcmd.NewDefaultClientConfig(*config, &overrides).ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create API client configuration from kubeconfig")
	}

	return kubernetes.NewForConfig(clientConfig)
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

	if cfg.MaxReleaseVersions == 0 {
		cfg.MaxReleaseVersions = 10
	}

	if cfg.TillerNS == "" {
		cfg.TillerNS = "kube-system"
	}

	if cfg.TillerLabel == "" {
		cfg.TillerLabel = "OWNER=TILLER"
	}

	// Build the label selector "OWNER=TILLER,NAME=myapp"
	cfg.TillerLabel += fmt.Sprintf(",NAME=%s", cfg.Release)

	convert.convertOptions = convertcmd.ConvertOptions{
		DeleteRelease:      cfg.DeleteV2Releases,
		DryRun:             cfg.DryRun,
		MaxReleaseVersions: cfg.MaxReleaseVersions,
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

// getV2ReleaseConfigmaps returns a list of configmaps that are helm v2 releases
func (c *Convert) getV2ReleaseConfigmaps(clientset kubernetes.Interface) (*corev1.ConfigMapList, error) {

	return clientset.CoreV1().ConfigMaps(c.convertOptions.TillerNamespace).List(metav1.ListOptions{
		LabelSelector: c.convertOptions.TillerLabel,
	})
}

// preserveV2ReleaseConfigmaps keeps the helm v2 configmaps by modifying a label
func (c *Convert) preserveV2ReleaseConfigmaps(clientset kubernetes.Interface, configmaps *corev1.ConfigMapList, ownerLabelValue string) error {

	tillerNamespace := c.convertOptions.TillerNamespace

	log.Printf("Preserving release versions of %s", c.convertOptions.ReleaseName)
	for _, item := range configmaps.Items {
		item.Labels["OWNER"] = ownerLabelValue

		if _, err := clientset.CoreV1().ConfigMaps(tillerNamespace).Update(&item); err != nil {
			return fmt.Errorf("Failure preserving release version %s", item.Name)
		}
	}

	return nil
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

	if !c.convertOptions.DeleteRelease {
		clientset, err := clientsetFromFile(c.kubeConfig)
		if err != nil {
			return err
		}

		configmaps, err := c.getV2ReleaseConfigmaps(clientset)
		if err != nil {
			return err
		}

		if err := convertcmd.Convert(c.convertOptions, kc); err != nil {
			return err
		}

		if err := c.preserveV2ReleaseConfigmaps(clientset, configmaps, "converted-to-helm3"); err != nil {
			return err
		}
	} else {
		if err := convertcmd.Convert(c.convertOptions, kc); err != nil {
			return err
		}
	}

	return nil
}

// Prepare checks required inputs
func (c *Convert) Prepare() error {

	if c.convertOptions.ReleaseName == "" {
		return fmt.Errorf("release is required")
	}

	return nil
}
