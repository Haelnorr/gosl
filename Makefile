# Makefile
.PHONY: build
.PHONY: migrate

BINARY_NAME=gosl

build:
	tailwindcss -i ./pkg/embedfs/files/css/input.css -o ./pkg/embedfs/files/css/output.css && \
	go mod tidy && \
	templ generate && \
	go generate ./cmd/gosl && \
	go build -ldflags="-w -s" -o ./bin/${BINARY_NAME}${SUFFIX} ./cmd/gosl

migrate:
	go mod tidy && \
	go generate ./cmd/migrate && \
	go build -ldsflags"-w -s" -o ./bin/migrate${SUFFIX} ./cmd/migrate

dev:
	templ generate --watch &\
	air &\
	tailwindcss -i ./pkg/embedfs/files/css/input.css -o ./pkg/embedfs/files/css/output.css --watch

run:
	make build
	bin/gosl

test:
	go mod tidy && \
	templ generate && \
	go generate ./cmd/gosl && \
	go test ./cmd/gosl
	go test ./pkg/db
	go test ./internal/middleware

tester:
	go mod tidy && \
	go run ./cmd/gosl --port 3232 --tester --loglevel trace
