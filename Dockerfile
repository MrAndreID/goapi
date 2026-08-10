# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.24

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

ENV CGO_ENABLED=0 \
    GOTOOLCHAIN=local \
    GOFLAGS=-mod=readonly

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download && \
    go mod verify

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=bind,source=cmd,target=cmd \
    --mount=type=bind,source=internal,target=internal \
    --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,id=go-build-${TARGETOS}-${TARGETARCH} \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
        -trimpath \
        -buildvcs=false \
        -ldflags="-s -w -buildid=" \
        -o /out/engine \
        ./cmd/api

FROM alpine:${ALPINE_VERSION} AS runtime

ARG APP_USER=goapi
ARG APP_UID=10001
ARG APP_PORT=10001
ARG ENV_FILE=.env.example

ENV TZ=Asia/Jakarta \
    APP_PORT=${APP_PORT}

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g ${APP_UID} -S ${APP_USER} && \
    adduser -u ${APP_UID} -G ${APP_USER} -S -H -D ${APP_USER} && \
    install -d -m 755 /app && \
    install -d -o ${APP_UID} -g ${APP_UID} -m 750 /app/storage /app/storage/log

WORKDIR /app

COPY --link --chown=${APP_UID}:${APP_UID} --chmod=400 ${ENV_FILE} ./.env

COPY --link --chown=0:0 --chmod=555 --from=builder /out/engine ./engine

USER ${APP_UID}:${APP_UID}

EXPOSE ${APP_PORT}

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD nc -z -w 2 127.0.0.1 "$APP_PORT" || exit 1

ENTRYPOINT ["/app/engine"]
