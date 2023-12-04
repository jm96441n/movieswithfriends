.DEFAULT_GOAL := setup
CONTAINER_NAME := movieswithfriendsdb

.PHONY: setuppostgres
setuppostgres:
	if [  ! "$(shell docker container inspect -f '{{.State.Running}}' $(CONTAINER_NAME))" = "true" ]; then docker run -d --rm --name $(CONTAINER_NAME) -p 5432:5432 -e POSTGRES_USER="user" -e POSTGRES_PASSWORD="password" -v ./_data:/var/lib/postgresql postgres:bullseye; fi

.PHONY: setuptls
setuptls:
	if [ ! -d "./tls" ]; then mkdir tls && cd tls && go run $(shell go env GOROOT)/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost; fi

.PHONY: migrate
migrate: setuppostgres
	cd migrations && GOOSE_DBSTRING="user=myuser password=password host=db sslmode=disable dbname=movies" GOOSE_DRIVER="postgres" goose up

.PHONY: setup
setup: setuppostgres setuptls migrate
