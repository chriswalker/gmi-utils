#!/usr/bin/env bash

lint() {
  echo "Running linters..."
  golangci-lint run ./...
}

build() {
  build-gmiget
  build-gmifmt
  build-gmilinks
}

build-gmiget() {
  echo "Building gmiget..."
  go build -o bin/gmiget cmd/gmiget/main.go
}

build-gmifmt() {
  echo "Building gmifmt..."
  go build -o bin/gmifmt cmd/gmifmt/main.go
}

build-gmilinks() {
  echo "Building gmilinks..."
  go build -o bin/gmilinks cmd/gmilinks/main.go
}

test() {
  echo "Running all tests..."
  go test -test.count=1 -cover ./...
}

help() {
  printf "%s <task> <args>\n" "$0"
  printf "Tasks:\n"
  compgen -A function | cat -n
}

die() {
  printf "%s\n" "$@"
  exit 1
}

action="$1"
case $action in
  lint | build | build-gmiget | build-gmifmt | build-gmilinks | build-gmisrv | test | help)
    "$@"
    ;;
  *)
    die "invalid action '${action}'"
    ;;
esac
