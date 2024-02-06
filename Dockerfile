FROM golang:1.21 as build

ARG PLUGIN_TYPE="drone"
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/drone-helm ./cmd/drone-helm

# --- Copy the cli to an image with helm already installed ---
FROM alpine/helm:3.13.3

COPY --from=build /go/bin/drone-helm /bin/
COPY --chmod=600 ./assets/kubeconfig.tpl /root/.kube/config.tpl

ENTRYPOINT [ "/bin/drone-helm" ]
