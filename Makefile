OUT_NAME := cheapgpt
BUILD_DEPS := $(shell find . -type f -not -name $(OUT_NAME))

all: build

build: $(OUT_NAME)

$(OUT_NAME): $(BUILD_DEPS)
	go build ./...

.PHONY: install
install:
	go install .

.PHONY: test
test:
	go test -v ./...
