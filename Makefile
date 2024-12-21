.DEFAULT_GOAL := run
CONTAINER_NAME := movieswithfriendsdb

.PHONY: setupprettier
setupprettier:
	npm install --save-dev prettier prettier-plugin-go-template

.PHONY: setuppostgres
setuppostgres:
	if [  ! "$(shell docker container inspect -f '{{.State.Running}}' $(CONTAINER_NAME))" = "true" ]; then docker run -d --rm --name $(CONTAINER_NAME) -p 5432:5432 -e POSTGRES_USER="user" -e POSTGRES_PASSWORD="password" -e POSTGRES_DB="movieswithfriends" -v ./_data:/var/lib/postgresql/data postgres:bullseye && sleep 5; fi

.PHONY: migrate
migrate: setuppostgres
	cd migrations && GOOSE_DBSTRING="user=$$DB_USERNAME password=$$DB_PASSWORD host=$$DB_HOST sslmode=disable dbname=movieswithfriends" GOOSE_DRIVER="postgres" goose up && cd ..

.PHONY: setup
setup: setuppostgres migrate

.PHONY: psql
psql: setuppostgres
	docker exec -it $(CONTAINER_NAME) psql -U user movieswithfriends

.PHONY: gen-session-key
gen-session-key:
	go run cmd/scripts/gen_session_key.go

.PHONY: run
run:
	air

.PHONY: build-deploy
build-deploy:
	go build -o ./infra/ansible/files/movieswithfriends ./cmd/movieswithfriends

.PHONY: copy-migrations
copy-migrations:
	cp -r ./migrations ./infra/ansible/files

.PHONY: deploy
deploy: build-deploy copy-migrations
	source ./infra/.envrc && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i './infra/ansible/inventory/app.yml' -u root --private-key ~/.ssh/do ./infra/ansible/deploy_app.yml
