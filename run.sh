#!/usr/bin/env bash

build() {
    build-gmiget
    build-gmifmt
}

build-gmiget() {
    echo "Building gmiget..."
    go build -o bin/gmiget cmd/gmiget/main.go
}

build-gmifmt() {
    echo "Building gmifmt..."
    go build -o bin/gmifmt cmd/gmifmt/main.go
}

test() {
    echo "Running all tests..."
    go test -cover ./...
}

help() {
    printf "%s <task> <args>\n" "$0"
    printf "Tasks:\n"
    compgen -A function | cat -n
}

die() {
    printf "%s\n" "$0"
    exit 1
}

action="$1"
case $action in
    build|build-gmiget|build-gmifmt|test|help)
        "$@"
    ;;
    *)
        die "invalid action '${action}"
    ;;
esac

