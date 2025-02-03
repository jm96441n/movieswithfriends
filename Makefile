### LOCAL DEV
.DEFAULT_GOAL := run

.PHONY: setupprettier
setupprettier:
	npm install --save-dev prettier prettier-plugin-go-template

.PHONY: migrate
migrate:
	cd migrations && GOOSE_DBSTRING="user=$$DB_USERNAME password=$$DB_PASSWORD host=$$DB_HOST sslmode=disable dbname=movieswithfriends" GOOSE_DRIVER="postgres" goose up && cd ..

.PHONY: build-assets
build-assets:
	go run ./tools/assetbuilder/

.PHONY: psql
psql:
	docker compose exec -it db psql -U user movieswithfriends

.PHONY: run
run:
	docker compose up --build

### TESTS
headless ?= true

.PHONY: e2e
e2e:
	cd ./e2e && gotestsum -- ./... -count=1 -run="$(TEST)" -headless=$(headless)

### INFRA/DEPLOY

.PHONY: pkr
pkr:
	cd ./infra/packer && packer build .

.PHONY: tf-plan
tf-plan:
	cd ./infra/terraform && terraform plan -out=tfplan

.PHONY: tf-apply
tf-apply: tf-plan
	cd ./infra/terraform && terraform apply "tfplan"

.PHONY: deploy
deploy:
	source ./infra/.envrc && kamal deploy
