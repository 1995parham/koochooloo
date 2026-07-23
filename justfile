default:
    @just --list

# build the embedded admin-panel SPA (web/dist is committed; rerun after UI changes)
web:
    @echo '{{ BOLD + CYAN }}Building admin panel{{ NORMAL }}'
    cd web && pnpm install && pnpm run build

# build koochooloo binary (uses the committed web/dist embed)
build:
    @echo '{{ BOLD + CYAN }}Building Koochooloo!{{ NORMAL }}'
    go build -o koochooloo ./cmd/koochooloo

# update go and frontend packages
update: update-web
    @cd ./cmd/koochooloo && go get -u

# update frontend packages to their latest versions
update-web:
    @echo '{{ BOLD + CYAN }}Updating admin panel packages{{ NORMAL }}'
    cd web && pnpm update --latest

# set up the dev environment with docker-compose
dev cmd *flags:
    #!/usr/bin/env bash
    echo '{{ BOLD + YELLOW }}Development environment based on docker-compose{{ NORMAL }}'
    set -eu
    set -o pipefail
    if [ {{ cmd }} = 'down' ]; then
      docker compose -f ./deployments/docker-compose.yml down --volumes --remove-orphans
    elif [ {{ cmd }} = 'up' ]; then
      docker compose -f ./deployments/docker-compose.yml up --wait -d {{ flags }}
    else
      docker compose -f ./deployments/docker-compose.yml {{ cmd }} {{ flags }}
    fi

# run tests in the dev environment
test: seed
    go test -race -v ./... -covermode=atomic -coverprofile=coverage.out

# point the CLI at the docker-compose postgres so migrate/seed exercise the SQL path
export koochooloo_database__dialect := "postgres"
export koochooloo_database__url := "host=127.0.0.1 user=koochooloo password=secret dbname=koochooloo port=5432 sslmode=disable"

seed: (dev "up")
    go run ./cmd/koochooloo/main.go migrate
    go run ./cmd/koochooloo/main.go seed

# connect into the dev environment database
database: (dev "up") (dev "exec" "database psql -U koochooloo koochooloo")

# run golangci-lint
lint:
    golangci-lint run -c .golangci.yml

k6: (dev "up") build
    ./koochooloo server > /dev/null 2>&1 &
    k6 run ./api/k6/script.js
