set dotenv-load

default:
    @just --list

# CSS
watch-css:
    npx tailwindcss -i ./base.css -o ./static/css/output.css --watch

# Database 
make-migration name:
	@goose -dir migrations $DATABASE $DATABASE_URL create {{name}} sql

generate-db:
	sqlc compile && sqlc generate --experimental

# Application
run:
    air -c .air.toml

