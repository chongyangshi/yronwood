#!/bin/sh

export YRONWOOD_LISTEN_ADDR="127.0.0.1:18080"
export YRONWOOD_STORAGE_DIRECTORY_PUBLIC="/tmp/yronwood_public"
export YRONWOOD_STORAGE_DIRECTORY_UNLISTED="/tmp/yronwood_unlisted"
export YRONWOOD_STORAGE_DIRECTORY_PRIVATE="/tmp/yronwood_private"
export YRONWOOD_AUTHENTICATION_BASIC_SECRET="5bhNT+ZIyxZuaxTIe1WFK1G5Su3YZfDDnOrBwrjts2c="
export YRONWOOD_AUTHENTICATION_BASIC_SALT="local-salt" # ^ = $(echo -n "local-secret:local-salt" | openssl dgst -binary -sha256 | openssl base64)

go run github.com/icydoge/yronwood