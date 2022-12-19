.PHONY: lint
lint: python-lint shell-lint
	echo "mock_uss lint complete"

.PHONY: python-lint
python-lint:
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:22.10.0 black --check . || (echo "Linter didn't succeed. You can use the following command to fix python linter issues: make format" && exit 1)

.PHONY: shell-lint
shell-lint:
	find . -name '*.sh' | xargs docker run --rm -v $(CURDIR):/code -w /code koalaman/shellcheck

.PHONY: format
format:
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:22.10.0 black .

.PHONY: test
test: lint
	TEST="true" ./run_locally_geoawareness.sh
