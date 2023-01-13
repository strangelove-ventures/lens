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

containerProtoVer=v0.2
containerProtoImage=tendermintdev/sdk-proto-gen:$(containerProtoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(containerProtoVer)
containerProtoGenSwagger=cosmos-sdk-proto-gen-swagger-$(containerProtoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(containerProtoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./scripts/protocgen.sh; fi
	@go mod tidy

.PHONY: all lint test race msan tools clean build
