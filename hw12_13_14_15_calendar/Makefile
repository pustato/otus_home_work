BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"
GRPC_GEN_DIR="internal/server/grpc/pb"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.release="develop" \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) \
	-X github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd.gitHash=$(GIT_HASH)

up: build-img
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.yaml up --remove-orphans

down:
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.yaml down

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

lint-fix: install-lint-deps
	golangci-lint run ./... --fix

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

mock:
	rm -rf internal/queue/mocks
	rm -rf internal/storage/mocks
	mockery --dir=internal/queue/. --all --output=internal/queue/mocks --packageprefix=mock
	mockery --dir=internal/storage/. --all --output=internal/storage/mocks --packageprefix=mock

dockerfile-lint:
	docker run --rm -i hadolint/hadolint < build/Dockerfile

integration-tests: build-img
	result_code=0 ;\
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.yaml -f ./deployments/docker-compose.test.yaml up --remove-orphans --build -d ;\
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.yaml -f ./deployments/docker-compose.test.yaml run test go test ./test/... || result_code=$$? ;\
	docker-compose --env-file=./deployments/.env -f ./deployments/docker-compose.yaml -f ./deployments/docker-compose.test.yaml down ;\
	exit $$result_code ;

.PHONY: up-dev down-dev build run build-img run-img version test install-lint-deps lint lint-fix generate evans mock dockerfile-lint integration-tests
