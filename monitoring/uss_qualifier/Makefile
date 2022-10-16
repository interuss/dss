.PHONY: test
test: validate_documentation lint
	./scripts/test_docker_fully_mocked.sh

.PHONY: validate_documentation
validate_documentation:
	./scripts/validate_test_definitions.sh

.PHONY: lint
lint:
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black --check . || (echo "Linter didn't succeed. You can use the following command to fix python linter issues: make format" && exit 1)
	find . -name '*.sh' | xargs docker run --rm -v $(CURDIR):/code -w /code koalaman/shellcheck

.PHONY: format
format:
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black .
