.PHONY: interuss
interuss:
	go get ./...

.PHONY: format
format:
	clang-format -style=file -i pkg/dssproto/dss.proto

.PHONY: install
install:
	cd $(shell mktemp -d) && go mod init tmp && go install golang.org/x/lint/golint

.PHONY: lint
lint: install
	golint ./...

pkg/dssproto/dss.pb.go: dss.proto
	protoc -I/usr/local/include -I.   -I$GOPATH/src   -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis   --go_out=plugins=grpc:. pkg/dssproto/dss.proto

pkg/dssproto/dss.pb.gw.go: dss.proto
	protoc -I/usr/local/include -I.   -I$GOPATH/src   -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis   --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. pkg/dssproto/dss.proto

pkg/dssproto/dss.proto: install-proto-generation
	openapi2proto -spec api.yaml -annotate > pkg/dssproto/dss.proto
	sed -i '' 's/package ds/package dssproto/g;s/service DSService/service DSServiceV0/g' pkg/dssproto/dss.proto 


.PHONY: install-proto-generation
install-proto-generation:
	go get -u github.com/NYTimes/openapi2proto/cmd/openapi2proto
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get -u github.com/golang/protobuf/protoc-gen-go


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