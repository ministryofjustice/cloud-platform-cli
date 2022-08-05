SOURCE_FILES := $(shell find * -name '*.go')

cloud-platform: $(SOURCE_FILES)
	go mod download
	go build -ldflags "-X github.com/ministryofjustice/cloud-platform-cli/pkg/commands.Version=Makefile" -o cloud-platform ./cmd/cloud-platform/main.go

test:
	go test ./...

fmt:
	go fmt ./...

.PHONY: test
