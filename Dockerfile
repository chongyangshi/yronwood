# ARG GO_ALPINE_VERSION=1.12.5-alpine3.9
# FROM golang:$GO_ALPINE_VERSION

# WORKDIR /go/src/github.com/icydoge/yronwood
# COPY . .

# RUN go get -d -v ./...
# RUN go install -v ./...

# CMD ["yronwood"]

# Dockerfile borrowed from https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/icydoge/yronwood/
COPY . .
RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/yronwood

FROM scratch
COPY --from=builder /go/bin/yronwood /go/bin/yronwood
ENTRYPOINT ["/go/bin/yronwood"]