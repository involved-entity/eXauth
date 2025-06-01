PROTO_AUTH_DIR := api/auth
PROTO_AUTH_FILE := $(PROTO_AUTH_DIR)/auth.proto

PROTO_USERS_DIR := api/users
PROTO_USERS_FILE := $(PROTO_USERS_DIR)/users.proto

GHZ_DATA_DIR := internal/pkg/ghz

generate-proto:
	protoc -I $(PROTO_AUTH_DIR) --go-grpc_out=. --go_out=. $(PROTO_AUTH_FILE)
	protoc -I $(PROTO_USERS_DIR) --go-grpc_out=. --go_out=. $(PROTO_USERS_FILE)

make-migrations:
	atlas migrate diff --env gorm

migrate:
	atlas migrate apply --env gorm

run:
	go run cmd/auth/main.go

run-test:
	@echo "Using CONFIG_PATH=/config/local.test.yml"
	@CONFIG_PATH=$(PWD)/config/local.test.yml go run cmd/auth/main.go

run-prod:
	@echo "Using CONFIG_PATH=/config/prod.yml"
	@docker-compose -f docker-compose.prod.yml up -d --build

test:
	@echo "Using CONFIG_PATH=/config/local.test.yml"
	@CONFIG_PATH=$(PWD)/config/local.test.yml go test -v -count=1 -p 1 ./tests/auth/... ./tests/users/...

docker-test:
	@echo "Using CONFIG_PATH=/config/local.docker.yml"
	@CONFIG_PATH=$(PWD)/config/local.docker.yml go test -v -count=1 -p 1 ./tests/auth/... ./tests/users/...

load-test:
	@export PATH=$PATH:$(go env GOPATH)/bin
	@echo "Starting auth.Auth.Register DDOS..."
	@ghz --insecure --proto ./$(PROTO_AUTH_FILE) --call auth.Auth.Register -n 1000 -c 5 --data-file ./$(GHZ_DATA_DIR)/register.json localhost:9090
	@echo "Starting auth.Auth.RegenerateCode DDOS..."
	@ghz --insecure --proto ./$(PROTO_AUTH_FILE) --call auth.Auth.RegenerateCode -n 1000 -c 5 --data-file ./$(GHZ_DATA_DIR)/regenerate_code.json localhost:9090
	@go run cmd/drop_users/main.go

docker-load-test:
	@export PATH=$PATH:$(go env GOPATH)/bin
	@echo "Starting auth.Auth.Register DDOS..."
	@ghz --insecure --proto ./$(PROTO_AUTH_FILE) --call auth.Auth.Register -n 1000 -c 5 --data-file ./$(GHZ_DATA_DIR)/register.json localhost:50051
	@echo "Starting auth.Auth.RegenerateCode DDOS..."
	@ghz --insecure --proto ./$(PROTO_AUTH_FILE) --call auth.Auth.RegenerateCode -n 1000 -c 5 --data-file ./$(GHZ_DATA_DIR)/regenerate_code.json localhost:50051
	@go run cmd/drop_users/main.go

.PHONY: generate, test
