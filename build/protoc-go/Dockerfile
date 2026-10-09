FROM golang:1.27

ARG PROTOC_VERSION=34.1
# Must match the google.golang.org/protobuf version in go.mod.
ARG PROTOC_GEN_GO_VERSION=v1.36.11

RUN apt-get update \
    && apt-get install -y --no-install-recommends unzip \
    && rm -rf /var/lib/apt/lists/*

RUN case "$(uname -m)" in \
        x86_64) arch=x86_64 ;; \
        aarch64|arm64) arch=aarch_64 ;; \
        *) echo "Unsupported architecture: $(uname -m)" && exit 1 ;; \
    esac \
    && curl -sSfL -o /tmp/protoc.zip \
        "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-${arch}.zip" \
    && unzip -q /tmp/protoc.zip -d /usr/local bin/protoc 'include/*' \
    && rm /tmp/protoc.zip

RUN GOBIN=/usr/local/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}

ENTRYPOINT ["protoc"]
