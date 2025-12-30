PACKAGES := $(shell go list ./...)

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## vet: vet code
.PHONY: vet
vet:
	go vet $(PACKAGES)

## test: run unit tests
.PHONY: test
test:
	go test -race -cover $(PACKAGES)

## build: build a binary
.PHONY: build
build: test
	go build -o ./app -v

## build-windows-amd64: build Windows AMD64 binary
.PHONY: build-windows-amd64
build-windows-amd64: test
	GOOS=windows GOARCH=amd64 go build -o ./mcp.exe -v

## build-windows-arm64: build Windows ARM64 binary
.PHONY: build-windows-arm64
build-windows-arm64: test
	GOOS=windows GOARCH=arm64 go build -o ./mcp-arm64.exe -v

## build-all: build binaries for all platforms
.PHONY: build-all
build-all: build build-windows-amd64 build-windows-arm64

## autobuild: auto build when source files change
.PHONY: autobuild
autobuild:
	# curl -sf https://gobinaries.com/cespare/reflex | sh
	reflex -g '*.go' -- sh -c 'echo "\n\n\n\n\n\n" && make build'

## dockerbuild: build project into a docker container image
.PHONY: dockerbuild
dockerbuild: test
	docker-compose build

## start: build and run local project
.PHONY: start
start: build
	clear
	@echo ""
	./app

## deploy: build code into a container and deploy it to the cloud dev environment
.PHONY: deploy
deploy: build
	./deploy.sh
