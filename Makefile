GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
COMMIT := $(shell scripts/git/commit.sh)
# LAST_RELEASE_TAG determines the version of the DSS and is baked into
# the executable using linker flags. We gracefully ignore any tag that 
# does not satisfy the naming pattern v*, thus supporting interleaving release
# and ordinary tags.
LAST_RELEASE_TAG := $(shell git describe --tags --abbrev=0 --match='v*' 2> /dev/null | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+')
LAST_RELEASE_TAG := $(or $(LAST_RELEASE_TAG), v0.0.0)

# Build and version information is baked into the executable itself.
BUILD_LDFLAGS := -X github.com/interuss/dss/pkg/build.time=$(shell date -u '+%Y-%m-%d.%H:%M:%S') -X github.com/interuss/dss/pkg/build.commit=$(COMMIT) -X github.com/interuss/dss/pkg/build.host=$(shell hostname)
VERSION_LDFLAGS := -X github.com/interuss/dss/pkg/version.tag=$(LAST_RELEASE_TAG) -X github.com/interuss/dss/pkg/version.commit=$(COMMIT)
LDFLAGS := $(BUILD_LDFLAGS) $(VERSION_LDFLAGS)

ifeq ($(OS),Windows_NT)
  detected_OS := Windows
else
  detected_OS := $(shell uname -s)
endif

.PHONY: interuss
interuss:
	go install -ldflags "$(LDFLAGS)" ./...

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

pkg/api/v1/ridpb/rid.pb.go: pkg/api/v1/ridpb/rid.proto
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --go_out=plugins=grpc:. $<

pkg/api/v1/ridpb/rid.pb.gw.go: pkg/api/v1/ridpb/rid.proto pkg/api/v1/ridpb/rid.pb.go
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. $<

pkg/api/v1/ridpb/rid.proto: install-proto-generation
	go run github.com/NYTimes/openapi2proto/cmd/openapi2proto \
		-spec interfaces/uastech/standards/remoteid/augmented.yaml -annotate \
		-out $@ \
		-tag dss \
		-indent 2 \
		-package ridpb

pkg/api/v1/auxpb/aux_service.pb.go:
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --go_out=plugins=grpc:. pkg/api/v1/auxpb/aux_service.proto

pkg/api/v1/auxpb/aux_service.pb.gw.go: pkg/api/v1/auxpb/aux_service.pb.go
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. pkg/api/v1/auxpb/aux_service.proto

pkg/api/v1/scdpb/scd.pb.go: pkg/api/v1/scdpb/scd.proto
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --go_out=plugins=grpc:. $<

pkg/api/v1/scdpb/scd.pb.gw.go: pkg/api/v1/scdpb/scd.proto pkg/api/v1/scdpb/scd.pb.go
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. $<

interfaces/scd_adjusted.yaml:
	./interfaces/adjuster/adjust_openapi_yaml.sh ./interfaces/astm-utm/Protocol/utm.yaml ./interfaces/scd_adjusted.yaml

pkg/api/v1/scdpb/scd.proto: interfaces/scd_adjusted.yaml install-proto-generation
# 	rm $(dir $@)
	go run github.com/NYTimes/openapi2proto/cmd/openapi2proto \
		-spec interfaces/scd_adjusted.yaml -annotate \
		-out $@ \
		-tag dss \
		-indent 2 \
		-package scdpb

.PHONY: install-proto-generation
install-proto-generation:
ifeq ($(shell which protoc),)
	$(error Proto generation requires that protoc be installed; please install protoc.  On a Mac: brew install protobuf  On Linux: See http://google.github.io/proto-lens/installing-protoc.html)
endif
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.14.3
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.14.3
	go get github.com/golang/protobuf/protoc-gen-go
ifeq ($(shell which protoc-gen-go),)
	$(error protoc-gen-go is not accessible after installation; GOPATH must be set and PATH must contain GOPATH/bin)
	# Example:
	# export GOPATH=/home/$USER/go
	# export PATH=$PATH:$GOPATH/bin
endif

.PHONY: protos
protos: pkg/api/v1/auxpb/aux_service.pb.gw.go pkg/api/v1/ridpb/rid.pb.gw.go pkg/api/v1/scdpb/scd.pb.gw.go;

.PHONY: install-staticcheck
install-staticcheck:
	go get honnef.co/go/tools/cmd/staticcheck

.PHONY: staticcheck
staticcheck: install-staticcheck
	staticcheck -go 1.12 ./...

.PHONY: test
test:
	go test -ldflags "$(LDFLAGS)" -count=1 -v ./...

.PHONY: test-cockroach
test-cockroach: cleanup-test-cockroach
	@docker run -d --name dss-crdb-for-testing -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v20.1.1 start --insecure > /dev/null
	go run ./cmds/db-manager/main.go --schemas_dir ./build/deploy/db-schemas/defaultdb --db_version v3.0.0 --cockroach_host localhost
	DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS=false go test -count=1 -v ./pkg/rid/store/cockroach -store-uri "postgresql://root@localhost:26257?sslmode=disable"
	DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS=false go test -count=1 -v ./pkg/scd/store/cockroach -store-uri "postgresql://root@localhost:26257?sslmode=disable"
	DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS=false go test -count=1 -v ./pkg/rid/application -store-uri "postgresql://root@localhost:26257?sslmode=disable"
	@docker stop dss-crdb-for-testing > /dev/null
	@docker rm dss-crdb-for-testing > /dev/null

.PHONY: cleanup-test-cockroach
cleanup-test-cockroach:
	@docker stop dss-crdb-for-testing > /dev/null 2>&1 || true
	@docker rm dss-crdb-for-testing > /dev/null 2>&1 || true

.PHONY: test-e2e
test-e2e:
	test/docker_e2e.sh
