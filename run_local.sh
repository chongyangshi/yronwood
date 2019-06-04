#!/bin/sh

LOCAL_SIGNING_KEY="-----BEGIN PRIVATE KEY-----
MHcCAQEEILb/QQXBXNRuHX6znGZgZpbUvrL5sKXRjqd15ETJJx7koAoGCCqGSM49
AwEHoUQDQgAEUGPrCkuMJVI32dsZz4e3gTYisAvcnd97cepQGuynsmUhW4AIQN0J
KUdlxCzAQZX2vlsQIzv9QdbX1gd4laatRA==
-----END PRIVATE KEY-----"

export YRONWOOD_LISTEN_ADDR="127.0.0.1:18080"
export YRONWOOD_INDEX_REDIRECT="https://google.co.uk"
export YRONWOOD_STORAGE_DIRECTORY_PUBLIC="/tmp/yronwood_public"
export YRONWOOD_STORAGE_DIRECTORY_UNLISTED="/tmp/yronwood_unlisted"
export YRONWOOD_STORAGE_DIRECTORY_PRIVATE="/tmp/yronwood_private"
export YRONWOOD_AUTHENTICATION_SIGHNING_KEY="${LOCAL_SIGNING_KEY}"
export YRONWOOD_AUTHENTICATION_BASIC_SECRET="e5b84d4fe648cb166e6b14c87b55852b51b94aedd865f0c39ceac1c2b8edb367"
export YRONWOOD_AUTHENTICATION_BASIC_SALT="local-salt" # ^ = $(echo -n "local-secret:local-salt" | openssl dgst -sha256 -hex)
export YRONWOOD_CORS_ALLOWED_ORIGIN="*"

go run github.com/icydoge/yronwood