.PHONY: lint
lint: validate-docs python-lint shell-lint

.PHONY: validate-docs
validate-docs:
	./scripts/validate_test_definitions.sh

.PHONY: python-lint
python-lint:
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black --check . || (echo "Linter didn't succeed. You can use the following command to fix python linter issues: make format" && exit 1)

.PHONY: shell-lint
shell-lint:
	find . -name '*.sh' | xargs docker run --rm -v $(CURDIR):/code -w /code koalaman/shellcheck

.PHONY: format
format: format-documentation
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black .

.PHONY: format-documentation
format-documentation:
	./scripts/format_test_documentation.sh

.PHONY: test
test: lint
	./scripts/test_docker_fully_mocked.sh
