BIN_DIR = $$(pwd)/bin
BINARY = $$(pwd)/bin/containerlab
MKDOCS_VER = 8.3.9
# insiders version/tag https://github.com/srl-labs/mkdocs-material-insiders/pkgs/container/mkdocs-material-insiders
# make sure to also change the mkdocs version in actions' cicd.yml and force-build.yml files
MKDOCS_INS_VER = 8.4.3-insiders-4.22.1

include .mk/lint.mk

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BINARY) -ldflags="-s -w -X 'github.com/srl-labs/containerlab/cmd.version=0.0.0' -X 'github.com/srl-labs/containerlab/cmd.commit=$$(git rev-parse --short HEAD)' -X 'github.com/srl-labs/containerlab/cmd.date=$$(date)'" main.go

build-with-podman:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -o $(BINARY) -ldflags="-s -w -X 'github.com/srl-labs/containerlab/cmd.version=0.0.0' -X 'github.com/srl-labs/containerlab/cmd.commit=$$(git rev-parse --short HEAD)' -X 'github.com/srl-labs/containerlab/cmd.date=$$(date)'" -trimpath -tags "podman exclude_graphdriver_btrfs btrfs_noversion exclude_graphdriver_devicemapper exclude_graphdriver_overlay containers_image_openpgp" main.go

test:
	go test -race ./... -v

MOCKDIR = ./mocks
.PHONY: mocks-gen
mocks-gen: mocks-rm ## Generate mocks for all the defined interfaces.
	go install github.com/golang/mock/mockgen@v1.6.0
	mockgen -package=mocks -source=nodes/node.go -destination=$(MOCKDIR)/node.go
	mockgen -package=mocks -source=clab/dependency_manager.go -destination=$(MOCKDIR)/dependency_manager.go
	mockgen -package=mocks -source=runtime/runtime.go -destination=$(MOCKDIR)/container_runtime.go

.PHONY: mocks-rm
mocks-rm: ## remove generated mocks
	rm -rf $(MOCKDIR)/*

lint:
	golangci-lint run

clint:
	docker run -it --rm -v $$(pwd):/app -w /app golangci/golangci-lint:v1.47.1 golangci-lint run --timeout 5m -v

.PHONY: docs
docs:
	docker run -v $$(pwd):/docs --entrypoint mkdocs squidfunk/mkdocs-material:$(MKDOCS_VER) build --clean --strict

.PHONY: serve
site:
	docker run -it --rm -p 8000:8000 -v $$(pwd):/docs squidfunk/mkdocs-material:$(MKDOCS_VER)

# serve the site locally using mkdocs-material insiders container
.PHONY: serve-insiders
serve-insiders:
	docker run -it --rm -p 8001:8000 -v $$(pwd):/docs ghcr.io/srl-labs/mkdocs-material-insiders:$(MKDOCS_INS_VER)

.PHONY: htmltest
htmltest:
	docker run --rm -v $$(pwd):/docs --entrypoint mkdocs squidfunk/mkdocs-material:$(MKDOCS_VER) build --clean --strict
	docker run --rm -v $$(pwd):/test wjdp/htmltest --conf ./site/htmltest-w-github.yml
	rm -rf ./site

# build containerlab bin and push it as an OCI artifact to ttl.sh and ghcr registries
# to obtain the pushed artifact use: docker run --rm -v $(pwd):/workspace ghcr.io/deislabs/oras:v0.11.1 pull ttl.sh/<image-name>
.PHONY: ttl-push
oci-push: build-with-podman
	@echo
	@echo "With the following pull command you get a containerlab binary at your working directory. To use this downloaded binary - ./containerlab deploy.... Make sure not forget to add ./ prefix in order to use the downloaded binary and not the globally installed containerlab!"
	@echo 'If https proxy is configured in your environment, pass the proxies via --env HTTPS_PROXY="<proxy-address>" flag of the docker run command.'
# push to ttl.sh
	docker run --rm -v $$(pwd)/bin:/workspace ghcr.io/oras-project/oras:v0.12.0 push ttl.sh/clab-$$(git rev-parse --short HEAD):1d ./containerlab
	@echo "download with: docker run --rm -v \$$(pwd):/workspace ghcr.io/oras-project/oras:v0.12.0 pull ttl.sh/clab-$$(git rev-parse --short HEAD):1d"
# push to ghcr.io
	@echo ""
	docker run --rm -v $$(pwd)/bin:/workspace -v $${HOME}/.docker/config.json:/root/.docker/config.json ghcr.io/oras-project/oras:v0.12.0 push ghcr.io/srl-labs/clab-oci:$$(git rev-parse --short HEAD) ./containerlab
	@echo "download with: docker run --rm -v \$$(pwd):/workspace ghcr.io/oras-project/oras:v0.12.0 pull ghcr.io/srl-labs/clab-oci:$$(git rev-parse --short HEAD)"
