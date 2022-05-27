.PHONY: ci
ci: install_dependencies test

.PHONY: install_dependencies
install_dependencies:
	pip install black

.PHONY: test
test: lint
	./scripts/test_docker_fully_mocked.sh

.PHONY: lint
lint:
	black --check .
	find . -name '*.sh' | xargs docker run --rm -v $(CURDIR):/code -w /code koalaman/shellcheck

.PHONY: format
format:
	black -v .
