USER_GROUP := $(shell id -u):$(shell id -g)
GOPATH := $(shell go env GOPATH 2> /dev/null)
GOBIN := $(GOPATH)/bin

UPSTREAM_OWNER := $(shell scripts/git/upstream_owner.sh)
COMMIT := $(shell scripts/git/commit.sh)
# DSS_VERSION_TAG determines the version of the DSS and is baked into
# the executable using linker flags. If the commit is not a tag,
# the version_tag will contain information about the closest tag
# (ie v0.0.1-g6a64c20, see RELEASE.md for more details).
DSS_VERSION_TAG := $(shell scripts/git/version.sh dss)


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
	cd monitoring/uss_qualifier && make format
	cd monitoring/monitorlib && make format
	gofmt -s -w .

.PHONY: lint
lint: go_lint shell_lint
	cd monitoring/uss_qualifier && make lint

.PHONY: go_lint
go_lint:
	docker run --rm -v $(CURDIR):/dss -w /dss golangci/golangci-lint:v1.50.1 golangci-lint run --timeout 5m --skip-dirs /dss/build/workspace --skip-files '.*\.gen\.go' -v -E gofmt,bodyclose,rowserrcheck,misspell,golint,staticcheck,vet

.PHONY: shell_lint
shell_lint:
	find . -name '*.sh' | grep -v '^./interfaces/astm-utm' | grep -v '^./build/workspace' | xargs docker run --rm -v $(CURDIR):/dss -w /dss koalaman/shellcheck

# --- Targets to autogenerate Go code for OpenAPI-defined interfaces ---
.PHONY: apis
apis: example_apis dummy_oauth_api dss_apis

openapi-to-go-server:
	docker image build -t interuss/openapi-to-go-server ./interfaces/openapi-to-go-server

dss_apis: openapi-to-go-server
	docker container run -u "$(USER_GROUP)" -it \
      	-v "$(CURDIR)/interfaces/aux/aux.yaml:/resources/auxv1.yaml" \
      	-v "$(CURDIR)/interfaces/astm-utm/Protocol/utm.yaml:/resources/scdv1.yaml" \
      	-v "$(CURDIR)/interfaces/rid/v1/remoteid/augmented.yaml:/resources/ridv1.yaml" \
        -v "$(CURDIR)/interfaces/rid/v2/remoteid/updated.yaml:/resources/ridv2.yaml" \
	    -v "$(CURDIR)/:/resources/src" \
			interuss/openapi-to-go-server \
		  		--api_import github.com/interuss/dss/pkg/api \
    	      	--api /resources/auxv1.yaml#dss \
    	      	--api /resources/scdv1.yaml#dss \
				--api /resources/ridv1.yaml#dss \
              	--api /resources/ridv2.yaml#dss@ridv2/rid/v2 \
    	      	--api_folder /resources/src/pkg/api

example_apis: openapi-to-go-server
	$(CURDIR)/interfaces/openapi-to-go-server/generate_example.sh

dummy_oauth_api: openapi-to-go-server
	docker container run -it \
		-v $(CURDIR)/interfaces/dummy-oauth/dummy-oauth.yaml:/resources/dummy-oauth.yaml \
		-v $(CURDIR)/cmds/dummy-oauth:/resources/output \
		interuss/openapi-to-go-server \
			--api_import github.com/interuss/dss/cmds/dummy-oauth/api \
			--api /resources/dummy-oauth.yaml \
			--api_folder /resources/output/api
# ---

.PHONY: install-staticcheck
install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck

.PHONY: staticcheck
staticcheck: install-staticcheck
	staticcheck -go 1.12 ./...

.PHONY: test
test:
	go test -ldflags "$(LDFLAGS)" -count=1 -v ./pkg/... ./cmds/...

.PHONY: test-cockroach
test-cockroach: cleanup-test-cockroach
	@docker run -d --name dss-crdb-for-testing -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v21.2.7 start-single-node --insecure > /dev/null
	go run ./cmds/db-manager/main.go --schemas_dir ./build/deploy/db_schemas/rid --db_version latest --cockroach_host localhost
	go test -count=1 -v ./pkg/rid/store/cockroach --cockroach_host localhost --cockroach_port 26257 cockroach_ssl_mode disable --cockroach_user root --cockroach_db_name rid --schemas_dir db-schemas/rid
	go test -count=1 -v ./pkg/scd/store/cockroach --cockroach_host localhost --cockroach_port 26257 cockroach_ssl_mode disable --cockroach_user root --cockroach_db_name scd --schemas_dir db-schemas/scd
	go test -count=1 -v ./pkg/rid/application --cockroach_host localhost --cockroach_port 26257 cockroach_ssl_mode disable --cockroach_user root --cockroach_db_name rid --schemas_dir db-schemas/rid
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
	scripts/tag.sh $(UPSTREAM_OWNER)/dss/$(VERSION)

start-locally:
	build/dev/run_locally.sh

stop-locally:
	build/dev/run_locally.sh stop
