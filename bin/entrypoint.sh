#!/bin/bash

MODE=$1

if [ -z "$MODE" ]; then
  echo "No mode set, defaulting to web"
  MODE="web"
fi

if [ "$MODE" = "web" ]; then
  echo "Running web server"
  /app/app-release-server
elif [ "$MODE" = "migrate" ]; then
  apt update
  apt install go
  go run github.com/steebchen/prisma-client-go migrate deploy
else
  echo "Running command $@"
  exec "$@"
fi
