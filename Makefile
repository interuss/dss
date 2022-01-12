GOPATH := $(shell go env GOPATH 2> /dev/null)
GOBIN := $(GOPATH)/bin

UPSTREAM_ORG := $(shell scripts/git/upstream_organization.sh)
COMMIT := $(shell scripts/git/commit.sh)
# DSS_VERSION_TAG determines the version of the DSS components and is baked into
# the executable using linker flags. If the commit is not a tag,
# the version_tag will contain information about the closest tag
# (ie v0.0.1-g6a64c20, see RELEASE.md for more details).
DSS_VERSION_TAG := $(shell scripts/git/version.sh dss)

GENERATOR_TAG := generator:${DSS_VERSION_TAG}

# Build and version information is baked into the executable itself.
BUILD_LDFLAGS := -X github.com/interuss/dss/pkg/build.time=$(shell date -u '+%Y-%m-%d.%H:%M:%S') -X github.com/interuss/dss/pkg/build.commit=$(COMMIT) -X github.com/interuss/dss/pkg/build.host=$(shell hostname)
VERSION_LDFLAGS := -X github.com/interuss/dss/pkg/version.tag=$(DSS_VERSION_TAG) -X github.com/interuss/dss/pkg/version.commit=$(COMMIT)
LDFLAGS := $(BUILD_LDFLAGS) $(VERSION_LDFLAGS)

ifeq ($(OS),Windows_NT)
  detected_OS := Windows
else
  detected_OS := $(shell uname -s)
endif

.PHONY: interuss
interuss:
	go install -ldflags "$(LDFLAGS)" ./cmds/...

go-mod-download: go.mod
	go mod download

go.mod:
	go mod tidy

.PHONY: format
format:
	clang-format -style=file -i pkg/api/v1/ridpb/rid.proto
	clang-format -style=file -i pkg/api/v1/scdpb/scd.proto
	clang-format -style=file -i pkg/api/v1/auxpb/aux_service.proto

.PHONY: lint
lint:
	docker run --rm -v $(CURDIR):/dss -w /dss golangci/golangci-lint:v1.26.0 golangci-lint run --timeout 5m -v -E gofmt,bodyclose,rowserrcheck,misspell,golint -D staticcheck,vet
	docker run --rm -v $(CURDIR):/dss -w /dss golangci/golangci-lint:v1.26.0 golangci-lint run --timeout 5m -v --disable-all  -E staticcheck --skip-dirs '^cmds/http-gateway,^pkg/logging'
	find . -name '*.sh' | grep -v '^./interfaces/astm-utm' | xargs docker run --rm -v $(CURDIR):/dss -w /dss koalaman/shellcheck

pkg/api/v1/ridpb/rid.pb.go: pkg/api/v1/ridpb/rid.proto generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) protoc \
		-I/usr/include \
		-I/src \
		-I/go/src \
		-I/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis \
		--go_out=plugins=grpc:. $<

pkg/api/v1/ridpb/rid.pb.gw.go: pkg/api/v1/ridpb/rid.proto pkg/api/v1/ridpb/rid.pb.go generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) protoc \
		-I/usr/include \
		-I. \
		-I/go/src \
		-I/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true,allow_delete_body=true:. $<

pkg/api/v1/ridpb/rid.proto: generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) openapi2proto \
		-spec interfaces/uastech/standards/remoteid/augmented.yaml -annotate \
		-tag dss \
		-indent 2 \
		-package ridpb > $@

pkg/api/v1/auxpb/aux_service.pb.go: pkg/api/v1/auxpb/aux_service.proto generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) protoc \
		-I/usr/include \
		-I. \
		-I/go/src \
		-I/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis \
		--go_out=plugins=grpc:. $<

pkg/api/v1/auxpb/aux_service.pb.gw.go: pkg/api/v1/auxpb/aux_service.proto pkg/api/v1/auxpb/aux_service.pb.go generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) protoc \
		-I/usr/include \
		-I. \
		-I/go/src \
		-I/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true,allow_delete_body=true:. $<

pkg/api/v1/scdpb/scd.pb.go: pkg/api/v1/scdpb/scd.proto generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) protoc \
		-I/usr/include \
		-I. \
		-I/go/src \
		-I/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis \
		--go_out=plugins=grpc:. $<

pkg/api/v1/scdpb/scd.pb.gw.go: pkg/api/v1/scdpb/scd.proto pkg/api/v1/scdpb/scd.pb.go generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) protoc \
		-I/usr/include \
		-I. \
		-I/go/src \
		-I/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true,allow_delete_body=true:. $<

interfaces/scd_adjusted.yaml: interfaces/astm-utm/Protocol/utm.yaml
	./interfaces/adjuster/adjust_openapi_yaml.sh ./interfaces/astm-utm/Protocol/utm.yaml ./interfaces/scd_adjusted.yaml

pkg/api/v1/scdpb/scd.proto: interfaces/scd_adjusted.yaml generator
	docker run -v$(CURDIR):/src:delegated -w /src $(GENERATOR_TAG) openapi2proto \
		-spec interfaces/scd_adjusted.yaml -annotate \
		-tag dss \
		-indent 2 \
		-package scdpb > $@

generator:
	docker build --rm -t $(GENERATOR_TAG) build/generator

.PHONY: protos
protos: pkg/api/v1/auxpb/aux_service.pb.gw.go pkg/api/v1/ridpb/rid.pb.gw.go pkg/api/v1/scdpb/scd.pb.gw.go

.PHONY: install-staticcheck
install-staticcheck:
	go get honnef.co/go/tools/cmd/staticcheck

.PHONY: staticcheck
staticcheck: install-staticcheck
	staticcheck -go 1.12 ./...

.PHONY: test
test:
	go test -ldflags "$(LDFLAGS)" -count=1 -v ./pkg/... ./cmds/...

.PHONY: test-cockroach
test-cockroach: cleanup-test-cockroach
	@docker run -d --name dss-crdb-for-testing -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v20.2.0 start-single-node --insecure > /dev/null
	go run ./cmds/db-manager/main.go --schemas_dir ./build/deploy/db_schemas/rid --db_version latest --cockroach_host localhost
	go test -count=1 -v ./pkg/rid/store/cockroach -store-uri "postgresql://root@localhost:26257?sslmode=disable"
	go test -count=1 -v ./pkg/scd/store/cockroach -store-uri "postgresql://root@localhost:26257?sslmode=disable"
	go test -count=1 -v ./pkg/rid/application -store-uri "postgresql://root@localhost:26257/rid?sslmode=disable"
	@docker stop dss-crdb-for-testing > /dev/null
	@docker rm dss-crdb-for-testing > /dev/null

.PHONY: cleanup-test-cockroach
cleanup-test-cockroach:
	@docker stop dss-crdb-for-testing > /dev/null 2>&1 || true
	@docker rm dss-crdb-for-testing > /dev/null 2>&1 || true

.PHONY: test-e2e
test-e2e:
	test/docker_e2e.sh

tag: VERSION = v$(MAJOR).$(MINOR).$(PATCH)

tag:
	scripts/tag.sh $(UPSTREAM_ORG)/dss/$(VERSION)

start-locally:
	build/dev/run_locally.sh

stop-locally:
	build/dev/run_locally.sh stop
