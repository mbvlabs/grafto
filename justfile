set dotenv-load

default:
    @just --list

# CSS
watch-css:
    npx tailwindcss -i ./resources/css/base.css -o ./static/css/output.css --watch

# Mails/MJML
compile-mails:
    ./node_modules/.bin/mjml -r ./resources/mails/*.mjml -o ./pkg/mail/templates/

watch-mail name:
    open ./pkg/mail/templates/{{name}}.html &
    ./node_modules/.bin/mjml --watch ./resources/mails/{{name}}.mjml -o ./pkg/mail/templates/{{name}}.html

# Database 
make-migration name:
	@goose -dir migrations $DATABASE $DATABASE_URL create {{name}} sql

generate-db:
	sqlc compile && sqlc generate --experimental

# Application
run:
    air -c .air.toml

