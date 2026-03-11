FROM ghcr.io/astral-sh/uv:python3.14-alpine

COPY --from=golang:1.22-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

ADD . /app
WORKDIR /app

ENV UV_LINK_MODE=copy
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --compile-bytecode

# Ensure UV don't check dependencies or lock file, since we installed everything before
ENV UV_FROZEN=1

# Ensure the cache folder is present and writable by anyone
# (Some command run with limited privileges)
RUN mkdir -p /.cache/uv/ && chmod 777 /.cache/uv/

ENTRYPOINT ["uv", "run", "generate.py"]
