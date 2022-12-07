.PHONY: lint
lint:
	cd uss_qualifier && make lint
	cd mock_uss && make lint
	cd monitorlib && make lint

.PHONY: python-lint
python-lint:
	cd uss_qualifier && make python-lint
	cd mock_uss && make python-lint
	cd monitorlib && make python-lint

.PHONY: shell-lint
shell-lint:
	cd uss_qualifier && make shell-lint
	cd mock_uss && make shell-lint

.PHONY: format
format:
	cd uss_qualifier && make format
	cd mock_uss && make format
	cd monitorlib && make format

.PHONY: build
build:
	./build.sh

.PHONY: test
test:
	cd monitorlib && make test
	cd mock_uss && make test
	cd uss_qualifier && make test
