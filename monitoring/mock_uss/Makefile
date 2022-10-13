.PHONY: lint
lint:
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black --check . || (echo "Linter didn't succeed. You can use the following command to fix python linter issues: make format" && exit 1)
	find . -name '*.sh' | xargs docker run --rm -v $(CURDIR):/code -w /code koalaman/shellcheck

.PHONY: format
format:
	docker run --rm -v $(CURDIR):/code -w /code pyfound/black:latest_release black .

.PHONY: test
test: lint
	TEST="true" ./run_locally_geoawareness.sh
