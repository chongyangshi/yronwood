.PHONY: build
SVC := yronwood
WEB_ALPINE_VERSION := 3.9
WEB_SVC := web-images-scy-email
COMMIT := $(shell git log -1 --pretty='%h')
REPOSITORY := 172.16.32.2:2443/go

.PHONY: pull build push

all: pull build push

web: web-pull web-build web-push

build:
	docker build -t ${SVC} ./Dockerfile

pull:
	docker pull golang:alpine

push:
	docker tag ${SVC}:latest ${REPOSITORY}:${SVC}-${COMMIT}
	docker push ${REPOSITORY}:${SVC}-${COMMIT}

web-build:
	docker build -t ${WEB_SVC} --build-arg ALPINE_VERSION=${WEB_ALPINE_VERSION} ./Dockerfile-web

web-pull:
	docker pull alpine:${WEB_ALPINE_VERSION}

web-push:
	docker tag ${WEB_SVC}:latest icydoge/web:${WEB_SVC}-${COMMIT}
	docker push icydoge/web:${WEB_SVC}-${COMMIT}