# ViComp — common tasks. Run `make` or `make help` for the list.

COMPOSE      ?= docker compose
APP_URL      ?= http://localhost:8090
DEV_DB_URL   ?= postgres://vicomp:vicomp@localhost:5433/vicomp?sslmode=disable
DEV_GOTENBERG?= http://localhost:3001

.DEFAULT_GOAL := help

## help: show this list
help:
	@echo "ViComp targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | awk -F': ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

## up: build + start the whole stack, then tail all logs (Ctrl-C stops tailing, not the stack)
up:
	$(COMPOSE) up --build -d
	@echo "app running at $(APP_URL) — live reload via air. Ctrl-C stops the log tail, not the containers."
	$(COMPOSE) logs -f

## up-fg: start the stack in the foreground (Ctrl-C stops the stack)
up-fg:
	$(COMPOSE) up --build

## down: stop the stack (keeps the database volume)
down:
	$(COMPOSE) down

## reset: stop the stack and delete the database volume (wipes all data)
reset:
	$(COMPOSE) down -v

## restart: recreate just the app container after a code change
restart:
	$(COMPOSE) up --build -d app
	@echo "app running at $(APP_URL)"

## rebuild: force a clean rebuild of the app image
rebuild:
	$(COMPOSE) build --no-cache app

## logs: follow logs for all services
logs:
	$(COMPOSE) logs -f

## logs-app: follow just the app logs
logs-app:
	$(COMPOSE) logs -f app

## ps: show container status
ps:
	$(COMPOSE) ps

## db: open a psql shell in the database container
db:
	$(COMPOSE) exec db psql -U vicomp -d vicomp

## dev: run db + gotenberg in Docker, run the Go app locally with reload-friendly output
dev:
	$(COMPOSE) up -d db gotenberg
	DATABASE_URL='$(DEV_DB_URL)' GOTENBERG_URL='$(DEV_GOTENBERG)' go run ./cmd/vicomp

## build: compile the binary to ./vicomp
build:
	go build -o vicomp ./cmd/vicomp

## tidy: sync go.mod / go.sum
tidy:
	go mod tidy

## fmt: gofmt the source tree
fmt:
	gofmt -w .

## vet: run go vet
vet:
	go vet ./...

## test: run the Go tests
test:
	go test ./...

.PHONY: help up up-fg down reset restart rebuild logs logs-app ps db dev build tidy fmt vet test
