SOURCE_FILES := $(shell find * -name '*.go')

cloud-platform: $(SOURCE_FILES)
	export GO111MODULE=on
	go mod download
	go build -o cloud-platform ./cmd/cloud-platform/main.go

test:
	go test ./...

fmt:
	go fmt ./...

.PHONY: test
