FROM alpine:3.4

RUN apk --no-cache add ca-certificates

COPY iplb-docker /iplb-docker

ENTRYPOINT ["/iplb-docker"]