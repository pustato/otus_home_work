BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"
GRPC_GEN_DIR="internal/server/grpc/pb"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.release="develop" \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.gitHash=$(GIT_HASH)

up-dev: build-img
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.dev.yaml up

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" .

run: build
	$(BIN) --config ./configs/config.yaml http

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

generate:
	rm -rf $(GRPC_GEN_DIR)
	mkdir -p $(GRPC_GEN_DIR)
	protoc \
			--proto_path=api/ \
			--go_out=$(GRPC_GEN_DIR) \
			--go-grpc_out=$(GRPC_GEN_DIR) \
			api/*.proto

evans:
	evans --proto=api/event_service.proto repl

.PHONY: build run build-img run-img version test lint up-dev generate
