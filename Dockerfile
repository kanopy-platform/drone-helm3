# --- Build the plugin cli first ---
FROM golang:1.15.2-alpine as builder

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/drone-helm ./cmd/drone-helm

# --- Copy the cli to an image with helm already installed ---
FROM alpine/helm:3.3.4

LABEL maintainer="MongoDB Infrastructure Team"
LABEL description="Helm v3 drone plugin with support for automatic migration from v2"
LABEL base="alpine/helm"

COPY --from=builder /go/bin/drone-helm /bin/drone-helm
COPY ./assets/kubeconfig.tpl /root/.kube/config.tpl

ENTRYPOINT [ "/bin/drone-helm" ]
