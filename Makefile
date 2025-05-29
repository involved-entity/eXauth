PROTO_AUTH_DIR := api/auth
PROTO_AUTH_FILE := $(PROTO_AUTH_DIR)/auth.proto

PROTO_USERS_DIR := api/users
PROTO_USERS_FILE := $(PROTO_USERS_DIR)/users.proto

CONFIG_PATH := $(PWD)/config/local.test.yml
export CONFIG_PATH

generate:
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

.PHONY: generate, test
