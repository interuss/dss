GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin

ifeq ($(OS),Windows_NT)
  detected_OS := Windows
else
  detected_OS := $(shell uname -s)
endif

ifeq ($(detected_OS),Windows)
	kubecfg_download := "unsupported"
endif
ifeq ($(detected_OS),Darwin)  # Mac OS X
	kubecfg_download := "https://github.com/bitnami/kubecfg/releases/download/v0.13.1/kubecfg-darwin-amd64"
endif
ifeq ($(detected_OS),Linux)
	kubecfg_download := "https://github.com/bitnami/kubecfg/releases/download/v0.13.1/kubecfg-linux-amd64"
endif

kubecfg_file := $(shell basename $(kubecfg_download))


.PHONY: interuss
interuss:
	go install -ldflags "-X github.com/interuss/dss/pkg/dss/build.time=$(shell date -u '+%Y-%m-%d.%H:%M:%S') -X github.com/interuss/dss/pkg/dss/build.commit=$(shell git rev-parse --short HEAD) -X github.com/interuss/dss/pkg/dss/build.host=$(shell hostname)" ./...

go-mod-download: go.mod
	go mod download

go.mod:
	go mod tidy

.PHONY: format
format:
	clang-format -style=file -i pkg/api/v1/dsspb/dss.proto
	clang-format -style=file -i pkg/api/v1/auxpb/aux.proto

.PHONY: install
install:
	cd $(shell mktemp -d) && go mod init tmp && go install golang.org/x/lint/golint

.PHONY: lint
lint: install
	golint ./...

pkg/api/v1/dsspb/dss.pb.go: pkg/api/v1/dsspb/dss.proto
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --go_out=plugins=grpc:. pkg/api/v1/dsspb/dss.proto

pkg/api/v1/dsspb/dss.pb.gw.go: pkg/api/v1/dsspb/dss.pb.go
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. pkg/api/v1/dsspb/dss.proto

pkg/api/v1/dsspb/dss.proto: install-proto-generation
	go run github.com/NYTimes/openapi2proto/cmd/openapi2proto \
		-spec api/v1/remoteid.yaml -annotate \
		-out pkg/dssproto/dss.proto \
		-tag dss \
		-indent 2 \
		-package dsspb

pkg/api/v1/auxpb/aux.pb.go: pkg/api/v1/auxpb/aux.proto
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --go_out=plugins=grpc:. pkg/api/v1/auxpb/aux.proto

pkg/api/v1/auxpb/aux.pb.gw.go: pkg/api/v1/auxpb/aux.pb.go
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.3/third_party/googleapis   --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. pkg/api/v1/auxpb/aux.proto

pkg/api/v1/auxpb/aux.proto: install-proto-generation
	go run github.com/NYTimes/openapi2proto/cmd/openapi2proto \
		-spec interfaces/v1/auxiliary.yaml -annotate \
		-out pkg/api/v1/auxpb/aux.proto \
		-tag dss \
		-indent 2 \
		-package auxpb

.PHONY: install-proto-generation
install-proto-generation:
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get github.com/golang/protobuf/protoc-gen-go

.PHONY: protos
protos: pkg/api/v1/auxpb/aux.pb.gw.go pkg/api/v1/dsspb/dss.pb.gw.go;

.PHONY: install-staticcheck
install-staticcheck:
	go get honnef.co/go/tools/cmd/staticcheck

.PHONY: kubecfg
kubecfg:
	mkdir -p temp
	wget $(kubecfg_download) -O ./temp/$(kubecfg_file)
	install ./temp/$(kubecfg_file) $(GOBIN)/kubecfg

.PHONY: staticcheck
staticcheck: install-staticcheck
	staticcheck -go 1.12 ./...

.PHONY: test
test: kubecfg
	go test -count=1 -v ./...

.PHONY: test-cockroach
test-cockroach: cleanup-test-cockroach
	@docker run -d --name dss-crdb-for-testing -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v19.1.2 start --insecure > /dev/null
	go test -count=1 -v ./pkg/dss/cockroach -store-uri "postgresql://root@localhost:26257?sslmode=disable"
	@docker stop dss-crdb-for-testing > /dev/null
	@docker rm dss-crdb-for-testing > /dev/null

.PHONY: cleanup-test-cockroach
cleanup-test-cockroach:
	@docker stop dss-crdb-for-testing > /dev/null 2>&1 || true
	@docker rm dss-crdb-for-testing > /dev/null 2>&1 || true

.PHONY: test-e2e
test-e2e:
	test/docker_e2e.sh
