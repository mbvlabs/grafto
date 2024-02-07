# Grafto - full-stack web dev in Go

The kickstarter repository for full-stack Go apps using your grandfather's technology.

Quick heads-up, this is still work in progress so expect lots of changes to come.

## Aim

The aim of Grafto is to be starter template that provides most of what you'll need to get a new web project off the 
ground, taking inspiration from Laravel equivalent. There are some opinionated choices made (like no ORM, old fashioned 
HTML rendered on the server) but tries to be as idiomatic as possible.

The target audience for the starter is mostly going to be solo-devs building side-projects or trying to bootstrap a 
new business.

It's important to note that there currently exists a much more feature complete starter template, called [pagado](https://github.com/mikestefanello/pagoda).
This is not an attempt at replacing that, but rather offer another approach and view to full-stack web development
in Go.

## Things left TODOs

- [ ] Refactor InternalError controller
- [ ] Update InternalError path's in controllers/user.go
- [ ] Add support for flash messages

## Getting started

Make sure you've the following dependencies installed:
- [Go v1.21+](https://golang.org/doc/install)
- [PostgreSQL v14+](https://www.postgresql.org/download/)
- [Goose](https://github.com/pressly/goose)
- [AIR](https://github.com/cosmtrek/air)
- [Justfile](https://github.com/casey/just)
- [Docker](https://docs.docker.com/get-docker/) | optional
- [sqlc 1.22.0](https://github.com/kyleconroy/sqlc)
- [templ](https://templ.guide/)

Next, run the cmd below to have your `.env` files ready:
```bash
cp .env.example .env
```

You'll need to setup a database before you can run the migrations, do that, and fill out the variable in the `.env` file
named "DATABASE_URL".

With that in-place, simply run:
```bash 
just migrations-up
``` 
and the database will be ready. Lastly, to run the application, open two terminals, run:
```bash 
just run
```
in one, and
```bash 
just watch-css
``` 
in another.

Visit `http://0.0.0.0:8000` to see the start page.
