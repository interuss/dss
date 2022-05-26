
.PHONY: test
test: lint
	./scripts/test_docker_fully_mocked.sh

.PHONY: lint
lint:
	find . -name '*.sh' | xargs docker run --rm -v $(CURDIR):/code -w /code koalaman/shellcheck

.PHONY: format
format:
	black -v .
