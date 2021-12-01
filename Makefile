VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
SDKVERSION := $(shell go list -m -u -f '{{.Version}}' github.com/cosmos/cosmos-sdk)
TMVERSION := $(shell go list -m -u -f '{{.Version}}' github.com/tendermint/tendermint)
COMMIT  := $(shell git log -1 --format='%H')

all: install

LD_FLAGS = -X github.com/jackzampolin/lens/cmd/cmd.Version=$(VERSION) \
	-X github.com/jackzampolin/lens/cmd/cmd.Commit=$(COMMIT) \
	-X github.com/jackzampolin/lens/cmd/cmd.SDKVersion=$(SDKVERSION) \
	-X github.com/jackzampolin/lens/cmd/cmd.TMVersion=$(TMVERSION)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

build:
	@go build -mod readonly $(BUILD_FLAGS) -o build/lens main.go

install:
	@go install -mod readonly $(BUILD_FLAGS) ./...

build-linux:
	@GOOS=linux GOARCH=amd64 go build --mod readonly $(BUILD_FLAGS) -o ./build/lens main.go

clean:
	rm -rf build

.PHONY: all lint test race msan tools clean build