ARG ALPINE_VERSION=3.9
FROM alpine:$ALPINE_VERSION
COPY ./yronwood /
CMD ["/yronwood"]