# Dockerfile borrowed from https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/chongyangshi/yronwood/
COPY . .
RUN GO111MODULE=off go get -d -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=off go build -ldflags="-w -s" -o /go/bin/yronwood

FROM scratch
COPY --from=builder /go/bin/yronwood /go/bin/yronwood
ENTRYPOINT ["/go/bin/yronwood"]
