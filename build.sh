#!/usr/bin/env bash

mkdir -p bin

go build -a -o bin/wait-host cmd/main.go
