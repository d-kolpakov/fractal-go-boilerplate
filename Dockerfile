FROM golang:1.14.6 as builder

RUN mkdir /app
ADD . /app/
WORKDIR /app


RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/service-entrypoint .
CMD ["./service-entrypoint"]
