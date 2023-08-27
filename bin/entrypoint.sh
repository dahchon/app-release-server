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
  apt install curl -y
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash - &&
    apt-get install -y nodejs

  npx -y prisma migrate deploy || exit 1
else
  echo "Running command $@"
  exec "$@"
fi
