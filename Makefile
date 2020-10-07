build:
	go build -o bin/drone-helm cmd/drone-helm/main.go

lint:
    golint -set_exit_status ./...

test:
	go test -cover ./...

sign:
	drone sign mongodb-forks/drone-helm3 --save