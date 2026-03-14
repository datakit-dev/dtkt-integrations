# syntax=docker/dockerfile:1
FROM cgr.dev/chainguard/go:latest-dev AS builder

ARG PKG_DIR=dtkt-integrations/packages/mailpit
WORKDIR /build

COPY --link ${PKG_DIR}/go.* ${PKG_DIR}/
COPY --link dtkt-sdk/sdk-go/go.* dtkt-sdk/sdk-go/

WORKDIR /build/${PKG_DIR}
RUN go mod edit -replace=github.com/datakit-dev/dtkt-sdk/sdk-go=../../../dtkt-sdk/sdk-go

RUN --mount=type=cache,target="/root/.cache/go-build" \
  --mount=type=cache,target="/root/go/pkg/mod"\
  go mod download

WORKDIR /build

COPY --link dtkt-sdk/sdk-go dtkt-sdk/sdk-go

COPY --link ${PKG_DIR}/package.dtkt.yaml ${PKG_DIR}/main.go ${PKG_DIR}/
COPY --link ${PKG_DIR}/pkg ${PKG_DIR}/pkg

RUN --mount=type=cache,target="/root/.cache/go-build" \
  --mount=type=cache,target="/root/go/pkg/mod"\
  CGO_ENABLED=0 \
  GOEXPERIMENT=jsonv2 \
  go build -C ${PKG_DIR} \
  -o ../../../bin/pkg \
  -ldflags="-s" .


FROM axllent/mailpit:v1.24.1 AS mailpit
FROM cgr.dev/chainguard/static:latest AS runner

COPY --link --from=mailpit /mailpit /bin/mailpit
COPY --link --from=builder /build/bin/pkg /bin/pkg

# gRPC, SMTP, POP3, HTTP
EXPOSE 9090/tcp 1025/tcp 1110/tcp 8025/tcp

ENTRYPOINT ["/bin/pkg"]
