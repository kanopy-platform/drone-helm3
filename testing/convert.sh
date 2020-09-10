#! /bin/sh

namespace="drone-helm3-convert-test"

# Temp files
tmpdir="/tmp/$(basename "$PWD")"
mkdir -p $tmpdir
secret_file="${tmpdir}/drone_helm3_convert_secret.env"
drone_pipeline="${tmpdir}/drone_helm3_convert_pipeline.yaml"

helm_setup() {
    echo ">>> Creating new namespace $namespace"
    kubectl create ns $namespace

    echo ">>> Configuring tiller RBAC"
    kubectl create -n $namespace serviceaccount tiller
    kubectl create -n $namespace rolebinding tiller-admin --clusterrole=admin --serviceaccount="${namespace}:tiller"

    local tiller_secret=`mktemp`

    cat <<EOF > $tiller_secret
apiVersion: v1
kind: Secret
type: kubernetes.io/service-account-token
metadata:
  annotations:
    kubernetes.io/service-account.name: tiller
  name: tiller-token
  namespace: $namespace
EOF
    kubectl apply -f $tiller_secret
    rm -rf $tiller_secret

    echo ">>> Running helm init"
    helm init --service-account tiller --tiller-namespace $namespace

    echo ">>> Waiting for tiller to be ready"
    kubectl wait -n $namespace --for=condition=Available deployment/tiller-deploy --timeout=30s
}

helm_install() {
    echo ">>> Installing example release"
    helm install --namespace $namespace --tiller-namespace $namespace -n example-1234 ./examples/mychart/
}

drone_secret_content() {
    echo "kubernetes_token=$(kubectl -n $namespace get secret tiller-token -ojsonpath='{.data.token}' | base64 --decode)"
}

drone_pipeline_content() {
    local image=$1
    local api_server=$2

    echo "kind: pipeline
type: docker
name: convert-test

steps:
  - name: convert-upgrade
    image: $image
    pull: always
    settings:
      mode: upgrade
      debug: true
      skip_tls_verify: true
      enable_v2_conversion: true
      delete_v2_releases: true
      chart: ./examples/mychart
      release: example-1234
      api_server: $api_server
      namespace: $namespace
      tiller_ns: $namespace
      kubernetes_token:
        from_secret: kubernetes_token"
}

drone_run() {

    local secret=$1
    local pipeline=$2
    local drone_cmd="drone exec --secret-file $secret $pipeline"

    echo ">>> Running drone exec. This cmd can be reused"; echo
    echo $drone_cmd; echo
    eval $drone_cmd
}

cleanup() {
    echo "Deleting namespace $namespace"
    kubectl delete ns $namespace

    echo "Cleaning up temporary files"
    rm -rf $tmpdir
}

while getopts ":hds" opt; do
  case ${opt} in
    h)
        echo "Usage:"
        echo "    convert.sh api_server drone-helm-image-uri  Runs drone exec to test convert plugin"
        echo "    convert.sh -h                               Display this help message."
        echo "    convert.sh -d                               Delete k8s namespace."
        echo "    convert.sh -s                               Setup helm2 requirements."
        echo "Example:"
        echo "    convert.sh localhost:5000/dh3 https://api.k8s.example.com"
        exit 0
        ;;
    d)
        cleanup
        exit 0
        ;;
    s)
        helm_setup
        helm_install
        exit 0
        ;;
   \?)
     echo "Invalid Option: -$OPTARG" 1>&2
     exit 1
     ;;
  esac
done
shift $((OPTIND -1))

# Main
echo "<<< Running drone >>>"
echo ">>> Drone secret with kubernetes_token."
echo
drone_secret_content | tee $secret_file
echo
echo ">>> Drone pipeline."
echo
drone_pipeline_content $1 $2 | tee $drone_pipeline
echo
drone_run $secret_file $drone_pipeline
