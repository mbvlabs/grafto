set dotenv-load

# alias
alias r := run
alias ra := run-app
alias rw := run-worker
alias re := run-email

alias ci := golangci

alias cm := create-migration
alias ms := migration-status
alias um := up-migrations
alias dm := down-migrations
alias dmt := down-migrations-to
alias rdb := reset-db

alias s := seed

alias gdf := generate-db-functions

alias ct := compile-templates
alias ft := fmt-templates

alias ex := explore

alias ti := test-integrations
alias tu := test-units

default:
    @just --list

# database 
create-migration name:
	@goose -dir psql/migrations postgres $DB_KIND://$DB_USER:$DB_PASSWORD@localhost:5432/$DB_NAME create {{name}} sql

migration-status:
	go run cmd/migration/main.go --cmd "status"

up-migrations:
	go run cmd/migration/main.go --cmd "up"

up-migrations-by-one:
	go run cmd/migration/main.go --cmd "upbyone"

down-migrations:
	go run cmd/migration/main.go --cmd "down"

down-migrations-to version:
	go run cmd/migration/main.go --cmd "down" --version {{version}}

fix-migrations:
	@goose -dir psql/migrations postgres $DB_KIND://$DB_USER:$DB_PASSWORD@localhost:5432/$DB_NAME fix

reset-db:
	go run cmd/migration/main.go --cmd "reset"

generate-db-functions:
	@sqlc compile && sqlc generate

# application
run:
    wgo -xdir views/emails -file=.js -file=.css -file=.go -file=.templ -xfile=_templ.go just compile-templates :: just run-app

run-app:
    go run cmd/app/main.go

# worker
run-worker:
    @go run ./cmd/worker/main.go

# emails
run-email:
    wgo -dir ./emails  -file=.go -file=.templ -xfile=_templ.go templ generate :: go run cmd/email/main.go

# assets
compile-templates:
    templ generate

fmt-templates:
    cd views && templ fmt .

# exploration
explore:
    @go run ./cmd/explore/main.go

seed:
	@go run ./cmd/seed/main.go

# code quality
golangci:
	golangci-lint run

vet:
	@go vet ./...

# testing
test-units:
	@go test -tags=unit ./...

test-integrations:
	@go test -tags=integration ./...

test-e2e:
	@go test -tags=e2e -v ./...

test-all:
	@go test ./...
