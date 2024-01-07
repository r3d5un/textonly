# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DSN} up

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## test: run all tests found
.PHONY: test
test:
	@echo 'Running tests'
	go test ./...

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	golines . -w
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

.PHONY: swagger
swagger:
	@echo 'Creating swagger documentation...'
	swag fmt
	swag init -g ./cmd/web/main.go

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/web: build the cmd/web application
.PHONY: build/web
build/web:
	@echo 'Creating swagger documentation...'
	swag init -g ./cmd/api/main.go
	@echo 'Building cmd/web...'
	go build -ldflags='-s' -o=./bin/web ./cmd/web
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/web ./cmd/web

## build/docker tag=$1: build the Docker image of the cmd/web application
.PHONY: build/docker
build/docker:
	@echo 'Creating swagger documentation...'
	swag init -g ./cmd/api/main.go
	@echo 'Building docker image with tag ${tag}'
	docker build -t textonly:${tag} .

.PHONY: build/docker-compose/db
build/docker-compose/db:
	make swagger
	make audit
	make vendor
	@echo 'Building and starting Docker Compose with "db" profile'
	docker compose --profile=db up --build


# ==================================================================================== #
# DEPLOY DIGITALOCEAN
# ==================================================================================== #

## digitalocean/deploy
.PHONY: digitalocean/deploy
digitalocean/deploy:
	@echo "Logging into DigitalOcean Registry"
	doctl registry login
	make audit
	make build/docker tag=latest
	@echo "Tagging build..."
	docker tag textonly:latest registry.digitalocean.com/r3d5un/textonly:latest
	@echo "Pushing to DigitalOcean"
	docker push registry.digitalocean.com/r3d5un/textonly:latest
