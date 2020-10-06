package run

import (
	"io/ioutil"
	"testing"

	"github.com/mongodb-forks/drone-helm3/internal/env"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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

func clientsetWithV2ConfigmapsMock() *fake.Clientset {

	return fake.NewSimpleClientset(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myapp.v1",
				Namespace: "example",
				Labels: map[string]string{
					"NAME":    "myapp",
					"OWNER":   "TILLER",
					"STATUS":  "DEPLOYED",
					"VERSION": "1",
				},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myapp.v2",
				Namespace: "example",
				Labels: map[string]string{
					"NAME":    "myapp",
					"OWNER":   "TILLER",
					"STATUS":  "DEPLOYED",
					"VERSION": "2",
				},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other.v1",
				Namespace: "example",
				Labels: map[string]string{
					"NAME":    "other",
					"OWNER":   "TILLER",
					"STATUS":  "DEPLOYED",
					"VERSION": "1",
				},
			},
		},
	)
}

func TestGetV2ReleaseConfigmaps(t *testing.T) {

	c := NewConvert(env.Config{Release: "myapp", TillerNS: "example"}, "", "")
	clientset := clientsetWithV2ConfigmapsMock()

	expectedConfigmaps := &corev1.ConfigMapList{
		Items: []corev1.ConfigMap{
			corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myapp.v1",
					Namespace: "example",
					Labels: map[string]string{
						"NAME":    "myapp",
						"OWNER":   "TILLER",
						"STATUS":  "DEPLOYED",
						"VERSION": "1",
					},
				},
			},
			corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myapp.v2",
					Namespace: "example",
					Labels: map[string]string{
						"NAME":    "myapp",
						"OWNER":   "TILLER",
						"STATUS":  "DEPLOYED",
						"VERSION": "2",
					},
				},
			},
		},
	}

	cmaps, err := c.getV2ReleaseConfigmaps(clientset)
	assert.NoError(t, err)
	assert.Equal(t, len(cmaps.Items), 2)
	assert.Equal(t, cmaps, expectedConfigmaps)
}

func TestPreserveV2ReleaseConfigmaps(t *testing.T) {

	c := NewConvert(env.Config{Release: "myapp", TillerNS: "example"}, "", "")
	clientset := clientsetWithV2ConfigmapsMock()

	cmaps, err := c.getV2ReleaseConfigmaps(clientset)
	assert.NoError(t, err)

	err = c.preserveV2ReleaseConfigmaps(clientset, cmaps, "none")
	assert.NoError(t, err)

	tests := []struct {
		namespace       string
		configmapName   string
		ownerLabelValue string
	}{
		{
			namespace:       "example",
			configmapName:   "myapp.v1",
			ownerLabelValue: "none",
		},
		{
			namespace:       "example",
			configmapName:   "myapp.v2",
			ownerLabelValue: "none",
		},
		{
			namespace:       "example",
			configmapName:   "other.v1",
			ownerLabelValue: "TILLER",
		},
	}

	for _, test := range tests {
		cm, err := clientset.CoreV1().ConfigMaps(test.namespace).Get(test.configmapName, metav1.GetOptions{})
		assert.NoError(t, err)
		assert.Equal(t, cm.Labels["OWNER"], test.ownerLabelValue)
	}
}
