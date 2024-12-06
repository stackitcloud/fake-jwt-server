FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY fake-jwt-server /fake-jwt-server

ENTRYPOINT ["/fake-jwt-server"]
