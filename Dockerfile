FROM node:22.6.0 AS build-resources

WORKDIR /

COPY resources resources
COPY views views
COPY package.json package.json
COPY package-lock.json package-lock.json
COPY vite.config.js vite.config.js

RUN npm ci
RUN npm run prod

FROM golang:1.24 AS build-go

ARG appRelease=0.0.1

ENV APP_RELEASE=$appRelease

WORKDIR /

COPY . .

COPY --from=build-resources static/css static/css

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=$APP_RELEASE" -mod=readonly -v -o app cmd/app/main.go

FROM scratch

WORKDIR /

COPY --from=build-go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-go app app
COPY --from=build-go static static 

CMD ["./app"]
