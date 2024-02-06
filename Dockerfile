FROM golang:1.21 as build

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/app ./cmd/drone-helm

# --- Copy the cli to an image with helm already installed ---
FROM alpine/helm:3.13.3

COPY --chmod=600 ./assets/kubeconfig.tpl /root/.kube/config.tpl
COPY --from=build /go/bin/app /

ENTRYPOINT [ "/app" ]
