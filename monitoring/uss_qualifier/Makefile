.PHONY: test
test: lint
	./scripts/test_docker_fully_mocked.sh

.PHONY: validate_documentation
validate_documentation:
	./scripts/validate_test_definitions.sh

.PHONY: lint
lint: validate_documentation
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black --check . || (echo "Linter didn't succeed. You can use the following command to fix python linter issues: make format" && exit 1)
	find . -name '*.sh' | xargs docker run --rm -v $(CURDIR):/code -w /code koalaman/shellcheck

.PHONY: format_documentation
format_documentation:
	./scripts/format_test_documentation.sh

.PHONY: format
format: format_documentation
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black .
