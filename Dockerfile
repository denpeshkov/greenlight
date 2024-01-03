# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22-rc

FROM golang:${GO_VERSION} AS build
ENV GOMODCACHE=/go/pkg/mod/
ENV GOCACHE=/.cache/go-build/
WORKDIR /src
COPY . .
RUN \
    --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=cache,target=${GOCACHE} \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    CGO_ENABLED=0 go build -o /go/bin/greenlight ./cmd

FROM scratch as greenlight
WORKDIR /app
COPY --from=build /go/bin/greenlight .
EXPOSE 8080
ENTRYPOINT [ "./greenlight", "-addr=:8080" ]