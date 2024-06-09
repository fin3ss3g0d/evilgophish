#!/usr/bin/env bash

# make our output look nice...
script_name="replace rid"

function print_good () {
    echo -e "[${script_name}] \x1B[01;32m[+]\x1B[0m $1"
}

function print_error () {
    echo -e "[${script_name}] \x1B[01;31m[-]\x1B[0m $1"
}

if [[ $# -ne 2 ]]; then
    print_error "Missing Parameters:"
    print_error "Usage:"
    print_error "./replace_rid <previous rid> <new rid>"
    print_error " - previous rid      - the previous rid value that was replaced"
    print_error " - new rid           - the new rid value to replace the previous"
    print_error "Example:"
    print_error "   ./replace_rid.sh user_id client_id"

    exit 2
fi

# Set variables from parameters
previous_rid="${1}"
new_rid="${2}"

function check_go () {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed! Exiting..."
        exit
    fi
}

function main () {
    check_go
    find . -type f -exec sed -i "s|${previous_rid}|${new_rid}|g" {} \;
    cd gophish || exit 1
    go build
    cd ..
    cd evilginx3 || exit 1
    go build
    cd ..
    print_good "Replaced previous rid and rebuilt successfully!"
}

main