PROTO_AUTH_DIR := api/auth
PROTO_AUTH_FILE := $(PROTO_AUTH_DIR)/auth.proto
CONFIG_PATH := $(PWD)/config/local.test.yml
export CONFIG_PATH

generate:
	protoc -I $(PROTO_AUTH_DIR) --go-grpc_out=. --go_out=. $(PROTO_AUTH_FILE)

make-migrations:
	atlas migrate diff --env gorm

migrate:
	atlas migrate apply --env gorm

run:
	go run cmd/auth/main.go

test:
	@echo "Using CONFIG_PATH=$(CONFIG_PATH)"
	@go test -v ./tests/...

.PHONY: generate, test
