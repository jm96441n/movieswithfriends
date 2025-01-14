.DEFAULT_GOAL := run

.PHONY: setupprettier
setupprettier:
	npm install --save-dev prettier prettier-plugin-go-template

.PHONY: migrate
migrate:
	cd migrations && GOOSE_DBSTRING="user=$$DB_USERNAME password=$$DB_PASSWORD host=$$DB_HOST sslmode=disable dbname=movieswithfriends" GOOSE_DRIVER="postgres" goose up && cd ..

.PHONY: psql
psql:
	docker compose exec -it db psql -U user movieswithfriends

.PHONY: run
run:
	docker compose up --build

.PHONY: build-deploy
build-deploy:
	go build -o ./infra/ansible/files/movieswithfriends ./cmd/movieswithfriends

.PHONY: copy-migrations
copy-migrations:
	cp -r ./migrations ./infra/ansible/files

.PHONY: deploy
deploy: build-deploy copy-migrations
	source ./infra/.envrc && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i './infra/ansible/inventory/app.yml' -u root --private-key ~/.ssh/do ./infra/ansible/deploy_app.yml

.PHONY: e2e
e2e:
	cd ./e2e && go test -count=1 ./...
