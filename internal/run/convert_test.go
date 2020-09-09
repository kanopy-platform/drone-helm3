package run

import (
	"io/ioutil"
	"testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/stretchr/testify/assert"
)

func mockActions(t *testing.T) *action.Configuration {
	a := &action.Configuration{}
	a.Releases = storage.Init(driver.NewMemory())
	a.KubeClient = &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}}
	a.Capabilities = chartutil.DefaultCapabilities
	a.Log = func(format string, v ...interface{}) {
		t.Logf(format, v...)
	}

	return a
}

func TestV3ReleaseFound(t *testing.T) {

	cfg := mockActions(t)

	opts := &release.MockReleaseOptions{
		Name: "myapp",
	}

	err := cfg.Releases.Create(release.Mock(opts))
	assert.NoError(t, err)

	assert.True(t, v3ReleaseFound("myapp", cfg))
	assert.False(t, v3ReleaseFound("doesnt_exists", cfg))
}
