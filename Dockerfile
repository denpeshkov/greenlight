# syntax=docker/dockerfile:1
ARG GO_VERSION=1.21
ARG GOLANGCI_LINT_VERSION=latest-alpine

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS base
ENV GOMODCACHE=/go/pkg/mod/
ENV GOCACHE=/.cache/go-build/
WORKDIR /src!
RUN --mount=type=bind,target=go.mod,source=go.mod \
    --mount=type=bind,target=go.sum,source=go.sum \
    go mod download -x

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=cache,target=${GOCACHE} \
    --mount=type=bind,target=. \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build ./...