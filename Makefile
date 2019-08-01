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

.PHONY: test
test:
	go test -count=1 -v ./...

.PHONY: test-cockroach
test-cockroach:
	@docker run -d --name dss-crdb-for-testing -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v19.1.2 start --insecure > /dev/null
	go test -count=1 -v ./pkg/dss/cockroach -store-uri "postgresql://root@localhost:26257?sslmode=disable"
	@docker stop dss-crdb-for-testing > /dev/null
	@docker rm dss-crdb-for-testing > /dev/null