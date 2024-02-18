FROM node:18.16.1 AS build-resources

ARG app_host=0.0.0.0:8000
ARG app_scheme=http
ARG is_ci=true

ENV CI=$is_ci
ENV APP_HOST=$app_host
ENV APP_SCHEME=$app_scheme

WORKDIR /

COPY resources/ resources
COPY views/ views
COPY pkg/mail pkg/mail
COPY static static

RUN cd resources && npm ci

RUN cd resources && npm run build-css
RUN cd resources && npm run build-mails

FROM golang:1.21 AS build-go

ENV CI=true
WORKDIR /

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .

RUN templ generate

COPY --from=build-resources static static
COPY --from=build-resources pkg/mail pkg/mail

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -mod=readonly -v -o app cmd/app/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -mod=readonly -v -o worker cmd/worker/main.go

FROM scratch

WORKDIR /

COPY --from=build-go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-go app app
COPY --from=build-go worker worker
COPY --from=build-go static static 
COPY --from=build-go resources/seo resources/seo

CMD ["app", "worker"]
