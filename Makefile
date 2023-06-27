GREEN='\033[0;32m'
C_ORANGE1='\033[38;5;214m'
NC='\033[0m'

help:
	@echo -e ${GREEN}build:"     "${C_ORANGE1}builds koochooloo binary${NC}
	@echo -e ${GREEN}dev-up:"    "${C_ORANGE1}setup the dev environment using docker compose${NC}
	@echo -e ${GREEN}dev-down:"  "${C_ORANGE1}tear down the dev environment using docker compose${NC}
	@echo -e ${GREEN}test:"      "${C_ORANGE1}run tests locally${NC}
	@echo -e ${GREEN}lint:"      "${C_ORANGE1}run lint locally${NC}
	@echo -e ${GREEN}update:"    "${C_ORANGE1}run \"go get -u\"${NC}

build: koochooloo

koochooloo:
	go build -o koochooloo ./cmd/koochooloo

update:
	@cd cmd/koochooloo && go get -u

dev-up:
	docker compose -f deployments/docker-compose.yml up -d

dev-down:
	docker compose -f deployments/docker-compose.yml down
	docker compose -f deployments/docker-compose.yml rm

dev-%:
	docker compose -f deployments/docker-compose.yml $*

test: dev-up
	go run cmd/koochooloo/main.go migrate
	go test -v ./... -covermode=atomic -coverprofile=coverage.out

lint:
	golangci-lint run -c .golangci.yml


.PHONY: help build update lint test dev-up dev-down
