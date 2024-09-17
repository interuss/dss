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
	gofmt -s -w .

.PHONY: lint
lint: python-lint shell-lint go-lint terraform-lint

.PHONY: check-hygiene
check-hygiene: python-lint hygiene shell-lint go-lint terraform-lint

.PHONY: python-lint
python-lint:
	echo "TODO: apply python-lint to interfaces/openapi-to-go-server"

.PHONY: hygiene
hygiene:
	test/repo_hygiene/repo_hygiene.sh

.PHONY: shell-lint
shell-lint:
	echo "===== Checking DSS shell lint =====" && find . -name '*.sh' | grep -v '^./interfaces/astm-utm' | grep -v '^./build/workspace' | xargs docker run --rm -v $(CURDIR):/dss -w /dss koalaman/shellcheck

.PHONY: go-lint
go-lint:
	echo "===== Checking Go lint (except for *.gen.go files) =====" && docker run --rm -v $(CURDIR):/dss -w /dss golangci/golangci-lint:v1.58.1 golangci-lint run --timeout 5m --skip-dirs /dss/build/workspace --skip-files '.*\.gen\.go' -v -E gofmt,bodyclose,rowserrcheck,misspell,staticcheck,vet

.PHONY: terraform-lint
terraform-lint:
	docker run --rm -w /opt/dss -v ./deploy:/opt/dss/deploy -e TF_LOG=TRACE hashicorp/terraform fmt -recursive -check
	./deploy/infrastructure/utils/generate_terraform_variables.sh --lint

# This mirrors the hygiene-tests continuous integration workflow job (.github/workflows/ci.yml)
.PHONY: hygiene-tests
hygiene-tests: check-hygiene

# --- Targets to autogenerate Go code for OpenAPI-defined interfaces ---
.PHONY: apis
apis: example_apis dummy_oauth_api dss_apis

openapi-to-go-server:
	docker image build -t interuss/openapi-to-go-server ./interfaces/openapi-to-go-server

dss_apis: openapi-to-go-server
	docker container run -u "$(USER_GROUP)" -it \
      	-v "$(CURDIR)/interfaces/aux_/aux_.yaml:/resources/auxv1.yaml" \
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

.PHONY: check-dss
check-dss: evaluate-tanka test-go-units test-go-units-crdb build-dss test-e2e

.PHONY: test-go-units
test-go-units:
	go test -ldflags "$(LDFLAGS)" -count=1 -v ./pkg/... ./cmds/...

.PHONY: test-go-units-crdb
test-go-units-crdb: cleanup-test-go-units-crdb
	@docker run -d --name dss-crdb-for-testing -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v24.1.3 start-single-node --insecure > /dev/null
	@until [ -n "`docker logs dss-crdb-for-testing | grep 'nodeID'`" ]; do echo "Waiting for CRDB to be ready"; sleep 3; done;
	go run ./cmds/db-manager/main.go --schemas_dir ./build/db_schemas/rid --db_version latest --cockroach_host localhost
	go test -count=1 -v ./pkg/rid/store/cockroach --cockroach_host localhost --cockroach_port 26257 --cockroach_ssl_mode disable --cockroach_user root --cockroach_db_name rid
	go test -count=1 -v ./pkg/rid/application --cockroach_host localhost --cockroach_port 26257 --cockroach_ssl_mode disable --cockroach_user root --cockroach_db_name rid
	@docker stop dss-crdb-for-testing > /dev/null
	@docker rm dss-crdb-for-testing > /dev/null

.PHONY: cleanup-test-go-units-crdb
cleanup-test-go-units-crdb:
	@docker stop dss-crdb-for-testing > /dev/null 2>&1 || true
	@docker rm dss-crdb-for-testing > /dev/null 2>&1 || true

.PHONY: build-dss
build-dss:
	build/dev/run_locally.sh build

.PHONY: test-e2e
test-e2e: down-locally build-dss start-locally probe-locally collect-local-logs down-locally

tag:
	scripts/tag.sh $(UPSTREAM_OWNER)/dss/v$(VERSION)

.PHONY: restart-all
restart-all: build-dss down-locally start-locally

.PHONY: start-locally
start-locally:
	build/dev/run_locally.sh up -d
	build/dev/wait_for_local_dss.sh

.PHONY: probe-locally
probe-locally:
	build/dev/probe_locally.sh

.PHONY: qualify-locally
qualify-locally:
	build/dev/qualify_locally.sh

.PHONY: collect-local-logs
collect-local-logs:
	docker logs dss_sandbox-local-dss-core-service-1 2> core-service-for-testing.log

.PHONY: stop-locally
stop-locally:
	build/dev/run_locally.sh stop

.PHONY: down-locally
down-locally:
	build/dev/run_locally.sh down

# This mirrors the dss-tests continuous integration workflow job (.github/workflows/ci.yml)
.PHONY: dss-tests
dss-tests: evaluate-tanka test-go-units test-go-units-crdb build-dss down-locally start-locally probe-locally collect-local-logs down-locally

.PHONY: evaluate-tanka
evaluate-tanka:
	docker container run -v $(CURDIR)/build/jsonnetfile.json:/build/jsonnetfile.json -v $(CURDIR)/build/deploy:/build/deploy grafana/tanka show --dangerous-allow-redirect /build/deploy/examples/minimum
	docker container run -v $(CURDIR)/build/jsonnetfile.json:/build/jsonnetfile.json -v $(CURDIR)/build/deploy:/build/deploy grafana/tanka show --dangerous-allow-redirect /build/deploy/examples/schema_manager

# This reproduces the entire continuous integration workflow (.github/workflows/ci.yml)
.PHONY: presubmit
presubmit: hygiene-tests dss-tests
