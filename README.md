# Grafto

## Aim

The aim of Grafto is to be starter template that provides most of what you'll need to get a new web project off the 
ground, taking inspiration from Laravel equivalent. There are some opinionated choices made (like no ORM, old fashioned 
HTML rendered on the server) but tries to be as idiomatic as possible.

The target audience for the starter is mostly going to be solo-devs building side-projects or trying to bootstrap a 
new business.

## Getting started

Make sure you've the following dependencies installed:
- [Go v1.21+](https://golang.org/doc/install)
- [PostgreSQL v14+](https://www.postgresql.org/download/)
- [Goose](https://github.com/pressly/goose)
- [AIR](https://github.com/cosmtrek/air)
- [Justfile](https://github.com/casey/just)
- [Docker](https://docs.docker.com/get-docker/) | optional
- [sqlc 1.20.0](https://github.com/kyleconroy/sqlc)

Next, run the cmd below to have your `.env` files ready:
```bash
    cp .env.example .env
```

You'll need to setup a database before you can run the migrations, do that, and fill out the variable in the `.env` file
named "DATABASE_URL".

With that in-place, simply run `just migrations-up` and the database will be ready. Lastly, to run the application,
open two terminals, run `just run` in one of them and `just watch-css` in another. Visit `http://0.0.0.0:8080` to see
the start page.

## Views
You can define `partials`, either using `unrolled/render`'s `partial_name-current_tmpl_name` or the one built in with
Go's template library, using `define`. A `define` can be reused throughout the templates by using either `template` or
`block`. Those two are effectively the same, but `block` lets you define a fallback. If you create a file under `partials/`
and put the content inside a `define`, you can use it anywhere by doing `template name`. (TODO look up why) Using 
`unrolled/render`, the `block` override only works when its defined inside a template __not__ in a layout file. I.e.
creating a `block` inside `layouts/base.html` will not be overridable. If you add a `block` to a `define` you can use 
that to add additional elements.
