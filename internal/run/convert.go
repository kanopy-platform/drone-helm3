package run

import (
	"fmt"
	"log"

	"github.com/helm/helm-2to3/pkg/v3"
	"github.com/helm/helm-2to3/pkg/common"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	utils "github.com/maorfr/helm-plugin-utils/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pelotech/drone-helm3/internal/env"
)

type Convert struct {
	*config
	release string

	deleteV2Releases bool
	dryRun bool
	kubeContext string
	kubeConfig string
	label string
	releaseVersionsMax int
	tillerNS string

	cmd cmd
}

func NewConvert(cfg env.Config, tillerNS string, kubeConfig string) *Convert {
	return &Convert{
		config: newConfig(cfg),
		release: cfg.Release,
		deleteV2Releases: cfg.DeleteV2Releases,
		dryRun: cfg.DryRun,
		kubeContext: cfg.Context,
		kubeConfig: kubeConfig,
		label: cfg.TillerLabel,
		releaseVersionsMax: cfg.ReleaseVersionsMax,
		tillerNS: tillerNS,
	}
}

// Execute runs the `2to3 convert` command from Prepare
// it detects if a conversion is needed and if not it will return no error
// if a conversion is required then any returned error is related to a step
// in the conversion process
func (c *Convert) Execute() error {

	clientSet, err := clientSetFromKubeConfig(c.kubeConfig)
	if err != nil {
		return err
	}

	// If Tiller is not present, utils.ListReleasesWithKubeConfig will panic
	if !isTillerPresent(clientSet, "name=tiller,app=helm") {
		log.Println("There is no Tiller Deployment running")
		return nil
	}

	// Check for existence of v2 Release (only if Tiller is running)
	var v2ReleaseFound bool

	listOptions := utils.ListOptions{
		ReleaseName: c.release,
		TillerNamespace: c.tillerNS,
		TillerLabel: c.label,
	}

	v2Releases, _ := utils.ListReleasesWithKubeConfig(listOptions, c.kubeConfig, c.kubeContext)
	if len(v2Releases) > 0 {
		log.Printf("A v2 Release of %s was found", c.release)
		v2ReleaseFound = true
	} else {
		log.Printf("There is no v2 Release of %s to migrate", c.release)
		return nil
	}

	// Check for existence of v3 Release
	var v3ReleaseFound bool

	kc := common.KubeConfig{
		Context: c.kubeContext,
		File:    c.kubeConfig,
	}

	v3cfg, err := v3.GetActionConfig(c.namespace, kc)
	if err != nil {
		return err
	}

	_, err = v3cfg.Releases.Deployed(c.release)
	if err != nil {
		log.Printf("No v3 Release of %s found", c.release)
	} else {
		log.Printf("A v3 Release of %s was found", c.release)
		v3ReleaseFound = true
	}

	if v2ReleaseFound && v3ReleaseFound {
		return fmt.Errorf("Release %s, has entries both in v2 and v3 format", c.release)
	}

	return c.cmd.Run()
}

// Prepare builds the 2to3 convert command
// 2to3 convert options:
//     --delete-v2-releases         v2 release versions are deleted after migration. By default, the v2 release versions are retained
//     --dry-run                    simulate a command
// -h, --help                       help for convert
//     --kube-context string        name of the kubeconfig context to use
//     --kubeconfig string          path to the kubeconfig file
// -l, --label string               label to select Tiller resources by (default "OWNER=TILLER")
// -s, --release-storage string     v2 release storage type/object. It can be 'secrets' or 'configmaps'. This is only used with the 'tiller-out-cluster' flag (default "secrets")
//     --release-versions-max int   limit the maximum number of versions converted per release. Use 0 for no limit (default 10)
// -t, --tiller-ns string           namespace of Tiller (default "kube-system")
//     --tiller-out-cluster         when  Tiller is not running in the cluster e.g. Tillerless
func (c *Convert) Prepare() error {

	args := c.globalFlags()
	args = append(args, "2to3", "convert")

	// This is a required flag but is always passed by drone-helm
	args = append(args, "--kubeconfig", c.kubeConfig)

	if c.deleteV2Releases {
		args = append(args, "--delete-v2-releases")
	}

	if c.dryRun {
		args = append(args, "--dry-run")
	}

	if c.kubeContext != "" {
		args = append(args, "--kube-context", c.kubeContext)
	}

	if c.label != "" {
		args = append(args, "--label", c.label)
	}

	if c.releaseVersionsMax != 0 {
		args = append(args, "--release-versions-max", string(c.releaseVersionsMax))
	}

	if c.tillerNS != "" {
		args = append(args, "--tiller-ns", c.tillerNS)
	}

	args = append(args, c.release)

	c.cmd = command(helmBin, args...)
	c.cmd.Stdout(c.stdout)
	c.cmd.Stderr(c.stderr)
	if c.debug {
		fmt.Fprintf(c.stderr, "Generated command: '%s'\n", c.cmd.String())
	}

	return nil
}

// isTillerPresent checks if a Deployment for Tiller is running
func isTillerPresent(clientset kubernetes.Interface, labelSelector string) bool {
	opts := metav1.ListOptions{LabelSelector: labelSelector}

	_, err := clientset.AppsV1().Deployments("").List(opts)
	if err != nil {
		return false
	}

	return true
}

// clientSetFromKubeConfig returns a kubernetes.Clientset from a KubeConfig file
func clientSetFromKubeConfig(kubeConfig string) (*kubernetes.Clientset, error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
