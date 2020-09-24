package run

import (
	"fmt"
	"log"

	convertcmd "github.com/helm/helm-2to3/cmd"
	"github.com/helm/helm-2to3/pkg/common"
	"github.com/pelotech/drone-helm3/internal/env"
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

	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create API client")
	}
	return client, nil
}

// getReleaseConfigmaps returns a list of configmaps that are helm v2 releases
func getReleaseConfigmaps(clientset kubernetes.Interface, release, tillerNamespace, tillerLabel string) (*corev1.ConfigMapList, error) {

	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	if tillerLabel == "" {
		tillerLabel = "OWNER=TILLER"
	}
	if release != "" {
		tillerLabel += fmt.Sprintf(",NAME=%s", release)
	}

	configMaps, err := clientset.CoreV1().ConfigMaps(tillerNamespace).List(metav1.ListOptions{
		LabelSelector: tillerLabel,
	})
	if err != nil {
		return nil, err
	}

	return configMaps, nil
}

// preserveV2 keeps the helm v2 configmaps by modifying a label
func preserveV2(clientset kubernetes.Interface, o convertcmd.ConvertOptions) error {

	configMaps, err := getReleaseConfigmaps(clientset, o.ReleaseName, o.TillerNamespace, o.TillerLabel)
	if err != nil {
		return err
	}

	cmClient := clientset.CoreV1().ConfigMaps(o.TillerNamespace)

	log.Printf("Preserving release versions of %s", o.ReleaseName)
	for _, item := range configMaps.Items {
		item.Labels["OWNER"] = "none"

		if _, err := cmClient.Update(&item); err != nil {
			return fmt.Errorf("Failure preserving release version %s", item.Name)
		}
	}

	return nil
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

	if err := convertcmd.Convert(c.convertOptions, kc); err != nil {
		return err
	}

	clientset, err := clientsetFromFile(c.kubeConfig)
	if err != nil {
		return err
	}

	if !c.convertOptions.DeleteRelease {
		if err := preserveV2(clientset, c.convertOptions); err != nil {
			return err
		}
	}

	return nil
}

// Prepare is not used but it's required to fulfill the Step interface
func (c *Convert) Prepare() error {
	return nil
}
