FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY fake-jwt-server /fake-jwt-server

ENTRYPOINT ["/fake-jwt-server"]
