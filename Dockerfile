FROM golang:1.24 AS build-go

ARG appRelease=0.0.1

ENV APP_RELEASE=$appRelease

WORKDIR /

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=$APP_RELEASE" -mod=readonly -v -o app cmd/app/main.go

FROM scratch

WORKDIR /

COPY --from=build-go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-go app app

CMD ["./app"]
