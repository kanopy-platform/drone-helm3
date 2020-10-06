module github.com/mongodb-forks/drone-helm3

go 1.13

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible

require (
	github.com/golang/mock v1.3.1
	github.com/helm/helm-2to3 v0.6.0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.2.8
	helm.sh/helm/v3 v3.1.3
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.2
	rsc.io/letsencrypt v0.0.3 // indirect
)
