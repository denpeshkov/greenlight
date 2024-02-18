.PHONY: all
all: help tidy lint test test/cover run

.PHONY: help
help: ## Display this help screen
	awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: ## Tidy
	go mod tidy -v
	go mod verify
	go fmt ./... 
	go vet ./..

.PHONY: lint
lint: ## Lint
	golangci-lint run

.PHONY: test
test: ## Test
	go test -v -race ./...

.PHONY: test/cover
test/cover: ## Test and cover
	go test -v -race -buildvcs -coverprofile=./coverage.out ./...
	go tool cover -html=./coverage.out

.PHONY: run
run: ## build and run app
	go vet ./...
	go build -race -o ./bin/greenlight ./cmd/greenlight
	./bin/greenlight -addr=:8080 -dsn=${POSTGRES_DSN} -token-secret=my_secret