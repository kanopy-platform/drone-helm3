# --- Build the plugin cli first ---
FROM golang:1.18-alpine as builder

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/drone-helm ./cmd/drone-helm

# --- Copy the cli to an image with helm already installed ---
FROM alpine/helm:3.8.1

COPY --from=builder /go/bin/drone-helm /bin/drone-helm
COPY ./assets/kubeconfig.tpl /root/.kube/config.tpl

ENTRYPOINT [ "/bin/drone-helm" ]
