VERSION := $(shell git describe --tags)
COMMIT  := $(shell git log -1 --format='%H')

all: test install

LD_FLAGS = -X github.com/strangelove-ventures/lens/cmd.Version=$(VERSION) \
	-X github.com/strangelove-ventures/lens/cmd.Commit=$(COMMIT) \

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

build:
	@echo "Building 20/20 vision"
	@go build -mod readonly $(BUILD_FLAGS) -o build/lens main.go

test:
	@go test -mod=readonly -race ./...

install:
	@echo "Installing Lens"
	@go install -mod readonly $(BUILD_FLAGS) ./...

build-linux:
	@GOOS=linux GOARCH=amd64 go build --mod readonly $(BUILD_FLAGS) -o ./build/lens main.go

clean:
	rm -rf build

.PHONY: all lint test race msan tools clean build
