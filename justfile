set dotenv-load

# alias
alias r := run-app
alias rw := run-worker
alias re := run-email

alias wc := watch-css

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

default:
    @just --list

# CSS
watch-css:
    npm run dev

# Database 
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

# Application
run-app:
    wgo -xdir ./views/emails -file=.go -file=.templ -xfile=_templ.go templ generate :: go run cmd/app/main.go

# Worker
run-worker:
    @go run ./cmd/worker/main.go

# Emails
run-email:
    wgo -dir ./emails -file=.txt -file=.go -file=.templ -xfile=_templ.go templ generate :: go run cmd/email/*.go

# templates
compile-templates:
    templ generate

fmt-templates:
    cd views && templ fmt .

# exploration
explore:
    @go run ./cmd/explore/main.go

seed:
	@go run ./cmd/seed/main.go
