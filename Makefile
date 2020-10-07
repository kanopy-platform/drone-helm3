build:
	go build -o bin/drone-helm cmd/drone-helm/main.go

.PHONY: lint
lint:
	golint -set_exit_status ./...

.PHONY: test
test:
	go test -cover ./...

.PHONY: sign
sign:
	drone sign mongodb-forks/drone-helm3 --save