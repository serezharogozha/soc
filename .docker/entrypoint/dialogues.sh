#!/bin/sh

if [[ ! -f .env]]; then
  echo "Creating .env file"
  cp .env.example .env
fi

go run ./cmd/main.go

eval "$@"