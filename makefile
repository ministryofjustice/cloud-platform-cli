TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
VERSION := $(TAG:v%=%)
ifneq ($(COMMIT), $(TAG_COMMIT))
	VERSION := $(VERSION)-next-$(COMMIT)-$(DATE)
endif
ifeq ($(VERSION),)
	VERSION := $(COMMIT)-$(DATE)
endif
ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

FLAGS := -ldflags "-X github.com/ministryofjustice/cloud-platform-cli/pkg/commands.Version=$(VERSION)"

build:
	go build $(FLAGS) -o cloud-platform-$(VERSION) .

run:
	go run $(FLAGS) .

test:
	go test -race -covermode=atomic -coverprofile=coverage.out -v ./...

integration-test:
	go test --tags integration -race -covermode=atomic -coverprofile=coverage.out -v ./...


fmt:
	go fmt ./...

.PHONY: test
