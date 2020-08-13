module github.com/pelotech/drone-helm3

go 1.13

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible

require (
	github.com/golang/mock v1.3.1
	github.com/helm/helm-2to3 v0.6.0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/maorfr/helm-plugin-utils v0.0.0-20200216074820-36d2fcf6ae86
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.17.9
	k8s.io/client-go v0.17.9
	rsc.io/letsencrypt v0.0.3 // indirect
)
