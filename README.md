# Drone plugin for Helm 3

[![Build Status](https://drone.corp.mongodb.com/api/badges/mongodb-forks/drone-helm3/status.svg)](https://drone.corp.mongodb.com/mongodb-forks/drone-helm3)
[![Go Report](https://goreportcard.com/badge/github.com/mongodb-forks/drone-helm3)](https://goreportcard.com/report/github.com/mongodb-forks/drone-helm3)

This plugin provides an interface between [Drone](https://drone.io/) and [Helm 3](https://github.com/kubernetes/helm):

* Lint your charts
* Deploy your service
* Delete your service
* Automatically migrate a release from helm v2 to v3

The plugin is inspired by [drone-helm](https://github.com/ipedrazas/drone-helm), which fills the same role for Helm 2. It provides a comparable feature-set and the configuration settings are backward-compatible.

**NOTE:** This a fork of [pelotech/drone-helm3](https://github.com/pelotech/drone-helm3) created with the main purpose of adding support to convert a v2 release to v3 as part of the plugin workflow.

## Example configuration

The examples below give a minimal and sufficient configuration for each use-case. For a full description of each command's settings, see [docs/parameter_reference.md](docs/parameter_reference.md).

### Linting

```yaml
steps:
  - name: lint
    image: quay.io/mongodb/drone-helm:v3
    settings:
      mode: lint
      chart: ./
```

### Installation and upgrade

```yaml
steps:
  - name: deploy
    image: quay.io/mongodb/drone-helm:v3
    settings:
      mode: upgrade
      chart: ./
      lint: false
      release: my-project
      # disable_v2_conversion: true
    environment:
      KUBE_API_SERVER: https://my.kubernetes.installation/clusters/a-1234
      KUBE_TOKEN:
        from_secret: kubernetes_token
```
**Note** that the `lint` settings field show in the example above defaults to `false`.

### Convert

```yaml
steps:
  - name: deploy
    image: quay.io/mongodb/drone-helm:v3
    settings:
      mode: convert
      chart: ./
      release: my-project
      namespace: my-namespace
      tiller_ns: tiller-namespace
      # delete_v2_releases: true
    environment:
      KUBE_API_SERVER: https://my.kubernetes.installation/clusters/a-1234
      KUBE_TOKEN:
        from_secret: kubernetes_token
```

### Uninstallation

```yaml
steps:
  - name: uninstall
    image: quay.io/mongodb/drone-helm:v3
    settings:
      mode: uninstall
      release: my-project
    environment:
      KUBE_API_SERVER: https://my.kubernetes.installation/clusters/a-1234
      KUBE_TOKEN:
        from_secret: kubernetes_token
```

## Upgrading from drone-helm

drone-helm3 is largely backward-compatible with drone-helm. There are some known differences:

* You'll need to migrate the deployments in the cluster [helm-v2-to-helm-v3](https://helm.sh/blog/migrate-from-helm-v2-to-helm-v3/).
  * Or automatically migrate v2 releases before upgrading by using the configured default values. 
  * Or use the standalone `mode: convert`.
* EKS is not supported. See [#5](https://github.com/pelotech/drone-helm3/issues/5) for more information.
* The `prefix` setting is no longer supported. If you were relying on the `prefix` setting with `secrets: [...]`, you'll need to switch to the `from_secret` syntax.
* During uninstallations, the release history is purged by default. Use `keep_history: true` to return to the old behavior.
* Several settings no longer have any effect. The plugin will produce warnings if any of these are present:
    * `purge` -- this is the default behavior in Helm 3
    * `recreate_pods`
    * `tiller_ns` -- Only used if `mode: convert` or `disable_v2_conversion == false` (default value)
    * `upgrade`
    * `canary_image`
    * `client_only`
    * `stable_repo_url`
* Several settings have been renamed, to clarify their purpose and provide a more consistent naming scheme. For backward-compatibility, the old names are still available as aliases. If the old and new names are both present, the updated form takes priority. Conflicting settings will make your `.drone.yml` harder to understand, so we recommend updating to the new names:
    * `helm_command` is now `mode`
    * `helm_repos` is now `add_repos`
    * `api_server` is now `kube_api_server`
    * `service_account` is now `kube_service_account`
    * `kubernetes_token` is now `kube_token`
    * `kubernetes_certificate` is now `kube_certificate`
    * `wait` is now `wait_for_upgrade`
    * `force` is now `force_upgrade`

Since helm 3 does not require Tiller, we also recommend switching to a service account with less-expansive permissions.

### [Contributing](docs/contributing.md)

This repo is setup in a way that if you enable a personal drone server to build your fork it will
 build and publish your image (makes it easier to test PRs and use the image till the contributions get merged)

* Build local ```DRONE_REPO_OWNER=josmo DRONE_REPO_NAME=drone-ecs drone exec```
* on your server (or cloud.drone.io) just make sure you have DOCKER_USERNAME, DOCKER_PASSWORD, and PLUGIN_REPO set as secrets
