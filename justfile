default:
  @just --list

# build koochooloo binary
build:
	go build -o koochooloo ./cmd/koochooloo

# update go packages
update:
	@cd ./cmd/koochooloo && go get -u

# set up the dev environment with docker-compose
dev-up:
	docker compose -f ./deployments/docker-compose.yml up -d

# tear down the dev environment
dev-down:
	docker compose -f ./deployments/docker-compose.yml down
	docker compose -f ./deployments/docker-compose.yml rm

# run tests in the dev environment
test: dev-up
	go run ./cmd/koochooloo/main.go migrate
	go test -v ./... -covermode=atomic -coverprofile=coverage.out

# run golangci-lint
lint:
	golangci-lint run -c .golangci.yml
