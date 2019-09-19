GOPATH := $(shell go env GOPATH)

.PHONY: interuss
interuss:
	go install -ldflags "-X github.com/steeling/InterUSS-Platform/pkg/dss/build.time=$(shell date -u '+%Y-%m-%d.%H:%M:%S') -X github.com/steeling/InterUSS-Platform/pkg/dss/build.commit=$(shell git rev-parse --short HEAD) -X github.com/steeling/InterUSS-Platform/pkg/dss/build.host=$(shell hostname)" ./...

.PHONY: format
format:
	clang-format -style=file -i pkg/dss_v1/dss.proto

.PHONY: install
install:
	cd $(shell mktemp -d) && go mod init tmp && go install golang.org/x/lint/golint

.PHONY: lint
lint: install
	golint ./...

pkg/dss_v1/dss.pb.go:
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.9.6/third_party/googleapis   --go_out=plugins=grpc:. pkg/dss_v1/dss.proto

pkg/dss_v1/dss.pb.gw.go:
	protoc -I/usr/local/include -I.   -I$(GOPATH)/src   -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.9.6/third_party/googleapis   --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. pkg/dss_v1/dss.proto

.PHONY: install-proto-generation
install-proto-generation:
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.9.6
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.9.6
	go get github.com/golang/protobuf/protoc-gen-go


.PHONY: test
test:
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