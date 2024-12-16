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

alias gdf := generate-db-functions

alias ct := compile-templates
alias ft := fmt-templates

alias rm := river-migrate-up

alias ex := explore

default:
    @just --list

# CSS
watch-css:
    @cd resources && npm run watch-css

# Database 
create-migration name:
	@goose -dir psql/migrations $DB_KIND $DATABASE_URL create {{name}} sql \
	&& goose -dir psql/migrations $DB_KIND $DATABASE_URL fix

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
	@goose -dir psql/migrations $DB_KIND $DATABASE_URL fix

reset-db:
	go run cmd/migration/main.go --cmd "reset"

generate-db-functions:
	sqlc compile && sqlc generate

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

# river
river-migrate-up:
	river migrate-up --database-url $QUEUE_DATABASE_URL

# exploration
explore:
    @go run ./cmd/explore/main.go

riverui:
	docker run -p 8080:8080 --env DATABASE_URL=$DATABASE_URL --network host ghcr.io/riverqueue/riverui:latest
