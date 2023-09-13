set dotenv-load

# alias
alias r := run

alias wc := watch-css

alias wm := watch-mail

alias mm := make-migration
alias um := up-migrations
alias dm := down-migrations
alias dmt := down-migrations-to
alias rdb := reset-db
alias gdf := generate-db-functions
alias mpts := copy-preline-to-static

default:
    @just --list

# CSS
watch-css:
    npx tailwindcss -i ./resources/css/base.css -o ./static/css/output.css --watch

# Preline
copy-preline-to-static:
    @cp -r ./node_modules/preline/dist/ ./static/js/preline

# Mails/MJML
compile-mails:
    ./node_modules/.bin/mjml -r ./resources/mails/*.mjml -o ./pkg/mail/templates/

watch-mail name:
    open ./pkg/mail/templates/{{name}}.html &
    ./node_modules/.bin/mjml --watch ./resources/mails/{{name}}.mjml -o ./pkg/mail/templates/{{name}}.html

# Database 
make-migration name:
	@goose -dir migrations $DATABASE_KIND $DATABASE_URL create {{name}} sql

up-migrations:
	@goose -dir migrations $DATABASE_KIND $DATABASE_URL up

down-migrations:
	@goose -dir migrations $DATABASE_KIND $DATABASE_URL down

down-migrations-to version:
	@goose -dir migrations $DATABASE_KIND $DATABASE_URL down-to {{version}}

reset-db:
	@goose -dir migrations $DATABASE_KIND $DATABASE_URL reset

generate-db-functions:
	sqlc compile && sqlc generate --experimental

# Application
run:
    air -c .air.toml

