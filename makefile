SOURCE_FILES := $(shell find * -name '*.go')
DIRS_WITH_TESTS := $(shell find * -type f -name '*_test.go' | xargs -n 1 dirname | sort | uniq)
DIRS_WITH_GOFILES := $(shell find * -type f -name '*.go' | xargs -n 1 dirname | sort | uniq)

cloud-platform: $(SOURCE_FILES)
	go mod download
	go build -o cloud-platform ./cmd/cloud-platform/main.go

test:
	go test ./...

fmt:
	go fmt ./...

.PHONY: test
