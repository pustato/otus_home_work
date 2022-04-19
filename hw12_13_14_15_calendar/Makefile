BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"
GRPC_GEN_DIR="internal/server/grpc/pb"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.release="develop" \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.gitHash=$(GIT_HASH)

.PHONY: up-dev
up-dev: build-img
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.dev.yaml up --remove-orphans


.PHONY: up-dev
down-dev:
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.dev.yaml down

.PHONY: build
build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" .

.PHONY: run
run: build
	$(BIN) --config ./configs/config.yaml http

.PHONY: build-img
build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

.PHONY: run-img
run-img: build-img
	docker run $(DOCKER_IMG)

.PHONY: version
version: build
	$(BIN) version

.PHONY: test
test:
	go test -race ./internal/...

.PHONY: install-lint-deps
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

.PHONY: lint
lint: install-lint-deps
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix: install-lint-deps
	golangci-lint run ./... --fix

.PHONY: generate
generate:
	rm -rf $(GRPC_GEN_DIR)
	mkdir -p $(GRPC_GEN_DIR)
	protoc \
			--proto_path=api/ \
			--go_out=$(GRPC_GEN_DIR) \
			--go-grpc_out=$(GRPC_GEN_DIR) \
			api/*.proto

.PHONY: evans
evans:
	evans --proto=api/event_service.proto repl

.PHONY: mock
mock:
	rm -rf internal/queue/mocks
	rm -rf internal/storage/mocks
	mockery --dir=internal/queue/. --all --output=internal/queue/mocks --packageprefix=mock
	mockery --dir=internal/storage/. --all --output=internal/storage/mocks --packageprefix=mock

.PHONY: dockerfile-lint
dockerfile-lint:
	docker run --rm -i hadolint/hadolint < build/Dockerfile
