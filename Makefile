.DEFAULT_GOAL := run
CONTAINER_NAME := movieswithfriendsdb

.PHONY: setupprettier
setupprettier:
	npm install --save-dev prettier prettier-plugin-go-template

.PHONY: setuppostgres
setuppostgres:
	if [  ! "$(shell docker container inspect -f '{{.State.Running}}' $(CONTAINER_NAME))" = "true" ]; then docker run -d --rm --name $(CONTAINER_NAME) -p 5432:5432 -e POSTGRES_USER="user" -e POSTGRES_PASSWORD="password" -e POSTGRES_DB="movieswithfriends" -v ./_data:/var/lib/postgresql/data postgres:bullseye && sleep 5; fi

.PHONY: setuptls
setuptls:
	if [ ! -d "./tls" ]; then mkdir tls && cd tls && go run $(shell go env GOROOT)/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost; fi

.PHONY: migrate
migrate: setuppostgres
	cd migrations && GOOSE_DBSTRING="user=$$DB_USERNAME password=$$DB_PASSWORD host=$$DB_HOST sslmode=disable dbname=movieswithfriends" GOOSE_DRIVER="postgres" goose up && cd ..

.PHONY: setup
setup: setuppostgres setuptls migrate

.PHONY: psql
psql: setuppostgres
	docker exec -it $(CONTAINER_NAME) psql -U user movieswithfriends

.PHONY: gen-session-key
gen-session-key:
	go run cmd/scripts/gen_session_key.go

.PHONY: run
run:
	air
