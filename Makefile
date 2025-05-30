PROTO_AUTH_DIR := api/auth
PROTO_AUTH_FILE := $(PROTO_AUTH_DIR)/auth.proto

PROTO_USERS_DIR := api/users
PROTO_USERS_FILE := $(PROTO_USERS_DIR)/users.proto

GHZ_DATA_DIR := internal/pkg/ghz

CONFIG_PATH := $(PWD)/config/local.test.yml
export CONFIG_PATH

generate-proto:
	protoc -I $(PROTO_AUTH_DIR) --go-grpc_out=. --go_out=. $(PROTO_AUTH_FILE)
	protoc -I $(PROTO_USERS_DIR) --go-grpc_out=. --go_out=. $(PROTO_USERS_FILE)

make-migrations:
	atlas migrate diff --env gorm

migrate:
	atlas migrate apply --env gorm

run:
	go run cmd/auth/main.go

test:
	@echo "Using CONFIG_PATH=$(CONFIG_PATH)"
	@go test -v -p 1 ./tests/auth/... ./tests/users/...

load-test:
	@export PATH=$PATH:$(go env GOPATH)/bin
	@echo "Starting auth.Auth.Register DDOS..."
	@ghz --insecure --proto ./$(PROTO_AUTH_FILE) --call auth.Auth.Register -n 1000 -c 5 --data-file ./$(GHZ_DATA_DIR)/register.json localhost:9090
	@echo "Starting auth.Auth.RegenerateCode DDOS..."
	@ghz --insecure --proto ./$(PROTO_AUTH_FILE) --call auth.Auth.RegenerateCode -n 1000 -c 5 --data-file ./$(GHZ_DATA_DIR)/regenerate_code.json localhost:9090
	@go run cmd/drop_users/main.go

.PHONY: generate, test
