# Use the official Golang image as the parent image
FROM golang:1.20.2-alpine

# Set the working directory to /pkg
WORKDIR /

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1

COPY ./ /app/soc

WORKDIR /app/soc/counter

RUN apk add --no-cache curl nano bash shadow postgresql-client build-base ca-certificates pkgconfig openssl-dev

ENV PKG_CONFIG_PATH /usr/lib/ssl

RUN mkdir /.cache
RUN chown nobody:nobody -R /.cache
RUN chown nobody:nobody -R /go/pkg
RUN chown nobody:nobody -R /go/pkg/mod

RUN go mod download

USER nobody
