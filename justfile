default:
    @just --list

# build koochooloo binary
build:
    @echo '{{ BOLD + CYAN }}Building Koochooloo!{{ NORMAL }}'
    go build -o koochooloo ./cmd/koochooloo

# update go packages
update:
    @cd ./cmd/koochooloo && go get -u

# set up the dev environment with docker-compose
dev cmd *flags:
    #!/usr/bin/env bash
    set -euxo pipefail
    @echo '{{ BOLD + YELLOW }}Development environment based on docker-compose{{ NORMAL }}'
    if [ {{ cmd }} = 'down' ]; then
      docker compose -f ./deployments/docker-compose.yml down
      docker compose -f ./deployments/docker-compose.yml rm
    elif [ {{ cmd }} = 'up' ]; then
      docker compose -f ./deployments/docker-compose.yml up --wait -d {{ flags }}
    else
      docker compose -f ./deployments/docker-compose.yml {{ cmd }} {{ flags }}
    fi

# run tests in the dev environment
test: seed
    go test -v ./... -covermode=atomic -coverprofile=coverage.out

seed: (dev "up")
    go run ./cmd/koochooloo/main.go migrate
    go run ./cmd/koochooloo/main.go seed

# connect into the dev environment database
database: (dev "up") (dev "exec" "database mongosh koochooloo")

# run golangci-lint
lint:
    golangci-lint run -c .golangci.yml

k6: (dev "up") build
    ./koochooloo server > /dev/null 2>&1 &
    k6 run ./api/k6/script.js
