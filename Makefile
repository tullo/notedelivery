SHELL = /bin/bash -o pipefail
export COMPOSE_DOCKER_CLI_BUILD = 1
export DOCKER_BUILDKIT = 1

build:
	@docker-compose build

up:
	@docker-compose up

down:
	@docker-compose down --remove-orphans

staticcheck:
	$$(go env GOPATH)/bin/staticcheck -go 1.16 -tests ./...

staticcheck-install:
	@GO111MODULE=on go install honnef.co/go/tools/cmd/staticcheck@v0.1.2
	@$$(go env GOPATH)/bin/staticcheck -debug.version
