GOPACKAGES?=$(shell find . -name '*.go' -not -path "./vendor/*" -exec dirname {} \;| sort | uniq)
GOFILES?=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
COVER_TEMP_FILE=/tmp/go-cover-infomodel-validation.tmp
BUILD_PATH=./cmd/app

all: help

.PHONY: help build run fmt vendor clean test coverage check vet lint simulations

help:
	@echo "fmt            - format application sources"
	@echo "test           - run tests"
	@echo "coverage       - run tests with coverage"

protoc:
	mkdir -p "api_pb"
	protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway \
		--grpc-gateway_out=logtostderr=true:./api_pb \
		--swagger_out=allow_merge=true,merge_file_name=api:. \
		--go_out=plugins=grpc:./api_pb ./proto/*.proto

fmt:
	go fmt $(GOPACKAGES)
clean:
	go clean

test: clean
	go test -v $(GOPACKAGES)

coverage: clean
	go test -v -cover $(GOPACKAGES)

coverage_show: clean
	go test -coverprofile=$COVER_TEMP_FILE $(GOPACKAGES) && go tool cover -html=$COVER_TEMP_FILE && unlink $COVER_TEMP_FILE

build: clean
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/service-entrypoint $(BUILD_PATH)