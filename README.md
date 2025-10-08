This project has been archived in favor of [Andurel](https://github.com/mbvlabs/andurel) - a rails-like framework in Go that was always the natural next step from this starter template.

# Grafto - full-stack web dev in Go

The kickstarter repository for full-stack Go apps using your grandfather's technology.

Quick heads-up, this is still work in progress so expect lots of changes to come.

Made by Morten, creator of the [Golang Blog Course](https://golangblogcourse.com?utm_source=github&utm_campaign=grafto).

## Aim

The aim of Grafto is to be starter template that provides most of what you'll need to get a new web project off the 
ground, taking inspiration from Laravel equivalent. There are some opinionated choices made (like no ORM, old fashioned 
HTML rendered on the server) but tries to be as idiomatic as possible.

The target audience for the starter is mostly going to be solo-devs building side-projects or trying to bootstrap a 
new business.

It's important to note that there currently exists a much more feature complete starter template, called [pagado](https://github.com/mikestefanello/pagoda).

This is not an attempt at replacing that, but rather offer another approach and view to full-stack web development
in Go.

## Usage

This section is rather empty right now but will be expanded upon soon. For now, here's the steps:
1. run `rename.sh`
2. `cp .env.example .env` and fill it out
3. run `just um` to apply migrations
4. run `just r` to run the app
