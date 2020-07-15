SOURCE_FILES := $(shell find * | grep -v cloud-platform)
DIRS_WITH_TESTS := $(shell find pkg cmd -type f -name '*_test.go' | xargs dirname | sort | uniq)

cloud-platform: $(SOURCE_FILES)
	go mod download
	go build -o cloud-platform ./cmd/cloud-platform/main.go

test:
	for dir in $(DIRS_WITH_TESTS); do (cd $${dir}; go test); done

.PHONY: test
