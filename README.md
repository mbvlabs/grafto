# Grafto - Full-Stack Web Development in Go

A modern, production-ready Go web application starter template that provides everything you need to build robust web applications quickly. Built with battle-tested technologies and modern development practices.

Created by MBV, creator of the [Golang Blog Course](https://golangblogcourse.com?utm_source=github&utm_campaign=grafto).

## Aim

The aim of Grafto is to be starter template that provides most of what you'll need to get a new web project off the 
ground, taking inspiration from Laravel equivalent. There are some opinionated choices made (like no ORM, old fashioned 
HTML rendered on the server) but tries to be as idiomatic as possible.

The target audience for the starter is mostly going to be solo-devs building side-projects or trying to bootstrap a 
new business.

This section is rather empty right now but will be expanded upon soon. For now, here's the steps:
1. run `rename.sh`
2. `cp .env.example .env` and fill it out
3. run `just um` to apply migrations
4. run `just r` to run the app