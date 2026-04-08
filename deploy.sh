#!/usr/bin/env bash

cd "$(dirname "$0")" || exit 1

# git pull --rebase origin main

go build -o green-api .
pkill green-api
sleep 1
./green-api &
disown
