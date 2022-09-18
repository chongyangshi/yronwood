.PHONY: build
SVC := yronwood
WEB_ALPINE_VERSION := 3.15
WEB_SVC := web-images-chongya-ng
COMMIT := $(shell git log -1 --pretty='%h')

.PHONY: pull build push

all: pull build push

web: web-pull web-build web-push

build:
	docker buildx build --platform linux/amd64 -t ${SVC} .

pull:
	docker pull golang:alpine

push:
	docker tag ${SVC}:latest icydoge/web:${SVC}-${COMMIT}
	docker push icydoge/web:${SVC}-${COMMIT}

clean:
	docker image prune -f

web-build:
	docker buildx build --platform linux/amd64 -t ${WEB_SVC} --build-arg ALPINE_VERSION=${WEB_ALPINE_VERSION} ./web

web-pull:
	docker pull alpine:${WEB_ALPINE_VERSION}

web-push:
	docker tag ${WEB_SVC}:latest icydoge/web:${WEB_SVC}-${COMMIT}
	docker push icydoge/web:${WEB_SVC}-${COMMIT}
