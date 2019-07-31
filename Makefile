.PHONY: interuss
interuss:
	go get ./...

.PHONY: format
format:
	clang-format -style=file -i pkg/dssproto/dss.proto
