default:
  @just --list

# build koochooloo binary
build:
  go build -o koochooloo ./cmd/koochooloo

# update go packages
update:
  @cd ./cmd/koochooloo && go get -u

# set up the dev environment with docker-compose
dev cmd *flags:
  #!/usr/bin/env bash
  set -euxo pipefail
  if [ {{cmd}} = 'down' ]; then
    docker compose -f ./deployments/docker-compose.yml down
    docker compose -f ./deployments/docker-compose.yml rm
  elif [ {{cmd}} = 'up' ]; then
    docker compose -f ./deployments/docker-compose.yml up -d {{flags}}
  else
    docker compose -f ./deployments/docker-compose.yml {{cmd}} {{flags}}
  fi

# run tests in the dev environment
test $koochooloo_telemetry__meter__enabled="false": (dev "up")
  go run ./cmd/koochooloo/main.go migrate
  go test -v ./... -covermode=atomic -coverprofile=coverage.out

# run golangci-lint
lint:
  golangci-lint run -c .golangci.yml
