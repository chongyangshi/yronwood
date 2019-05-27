.PHONY: build
SVC := yronwood
COMMIT := $(shell git log -1 --pretty='%h')
REPOSITORY := 172.16.32.2:2443/go

.PHONY: pull build push

all: pull build push

build:
	go build -ldflags "-s -w" github.com/icydoge/yronwood
	docker build -t ${SVC} .

pull:
	docker pull golang:latest

push:
	docker tag ${SVC}:latest ${REPOSITORY}:${SVC}-${COMMIT}
	docker push ${REPOSITORY}:${SVC}-${COMMIT}