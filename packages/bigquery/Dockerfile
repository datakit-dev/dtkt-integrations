# syntax=docker/dockerfile:1
FROM cgr.dev/chainguard/go:latest-dev AS builder

ARG PKG_DIR=dtkt-integrations/packages/bigquery
WORKDIR /build

# ZetaSQL requires CGO and specifies clang as the C compiler
RUN apk add --no-cache clang

COPY --link ${PKG_DIR}/go.* ${PKG_DIR}/
COPY --link dtkt-sdk/sdk-go/go.* dtkt-sdk/sdk-go/

WORKDIR /build/${PKG_DIR}

RUN --mount=type=cache,target="/root/.cache/go-build" \
  --mount=type=cache,target="/root/go/pkg/mod"\
  go mod download

WORKDIR /build

COPY --link dtkt-sdk/sdk-go dtkt-sdk/sdk-go

COPY --link ${PKG_DIR}/package.dtkt.yaml ${PKG_DIR}/main.go ${PKG_DIR}/
COPY --link ${PKG_DIR}/pkg ${PKG_DIR}/pkg

RUN --mount=type=cache,target="/root/.cache/go-build" \
  --mount=type=cache,target="/root/go/pkg/mod"\
  CXX=clang++ \
  CGO_ENABLED=1 \
  GOEXPERIMENT=jsonv2 \
  go build -C ${PKG_DIR} \
  -o ../../../bin/pkg \
  -ldflags='-s -extldflags "-static"' .

FROM cgr.dev/chainguard/static:latest AS runner

COPY --link --from=builder /build/bin/pkg /bin/pkg

# gRPC
EXPOSE 9090/tcp

ENTRYPOINT ["/bin/pkg"]
