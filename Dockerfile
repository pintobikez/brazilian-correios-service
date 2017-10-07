FROM alpine:3.5

LABEL maintainer "pinto.bikez@gmail.com"

ARG APP_NAME=brazilian-correios-service

RUN apk add --no-cache ca-certificates

ADD ./build/$APP_NAME /app
ADD ./core.database.yml.example /core.database.yml
ADD ./core.correios.yml.example /core.correios.yml

# Environment Variables
ENV LISTEN "0.0.0.0:8080"
ENV DATABASE_FILE "core.database.yml"
ENV CORREIOS_FILE "core.correios.yml"
ENV LOG_FOLDER "/var/log/"

CMD ["/app"]
