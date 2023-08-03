#!make
include .env
export

pgmigrations := $(wildcard ./database/postgres/migrations/*.sql)
protobuf := $(wildcard ./**/*.proto)

.PHONY: tools test

all: gen

# generate sql builder files for type safe SQL
gen: $(pgmigrations)
	@echo Detected change in postgres migrations, generating new SQL types
# remove entire dir so make knows it has been updated
	rm -rf /.gen
	jet -dsn=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable -schema=${POSTGRES_SCHEMA} -path=./gen

gengrpc: $(protobuf)
	@echo Detected change in protobuf files, generating protobuf and gRPC types
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative **/pb/*.proto

# migrate postgres 1 step down
pgdown:
	migrate -path database/postgres/migrations -database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable down 1
# migrate postgres 1 step up
pgup:
	migrate -path ./database/postgres/migrations -database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable up 1

build:
	go build -tags RELEASE -o ./dist/rp_tracker ./cmd/tracker
	go build -tags RELEASE -o ./dist/rcaptv ./cmd/rcaptv

build_dev:
	go build -o ./dist/rp_tracker ./cmd/tracker
	go build -o ./dist/rcaptv ./cmd/rcaptv

start:
	docker compose up --build
start_tracker:
	docker compose up pg_db tracker --build
start_web:
	docker compose up pg_db rcaptv --build

test:
	go clean -testcache
	go test -timeout 30s ./...

tools:
# Jet cli for generating types for SQL
	@echo Installing go-jet
	go install github.com/go-jet/jet/v2/cmd/jet@latest
# go migrate
	@echo Installing go-migrate
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo Installing protoc-gen and grpc-gen
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	export PATH="$PATH:$(go env GOPATH)/bin"
	@echo Install protoc by your own please depending on your OS:
	@echo Linux
	@echo $ apt install -y protobuf-compiler
	@echo MacOS with homebrew
	@echo $ brew install protobuf
