# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22-rc
ARG GOLANGCI_LINT_VERSION=v1.55.2

FROM golang:${GO_VERSION} AS deps
ENV GOMODCACHE=/go/pkg/mod/
ENV GOCACHE=/.cache/go-build/
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=cache,target=${GOCACHE} \
    go mod download -x

FROM deps AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=cache,target=${GOCACHE} \
    CGO_ENABLED=0 go build -o /go/bin/greenlight ./cmd

FROM scratch as greenlight
WORKDIR /app
COPY --from=build /go/bin/greenlight .
EXPOSE 8080
ENTRYPOINT [ "./greenlight", "-addr=:8080" ]

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION} as lint
WORKDIR /lint
COPY . .
RUN golangci-lint run -v ./...