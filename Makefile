PROTO_DIR := api
PROTO_FILE := $(PROTO_DIR)/main.proto
CONFIG_PATH := $(PWD)/config/local.test.yml
export CONFIG_PATH

generate:
	protoc -I $(PROTO_DIR) --go-grpc_out=. --go_out=. $(PROTO_FILE)

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
