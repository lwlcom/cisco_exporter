FROM golang:alpine AS builder
RUN apk add git
RUN go get -u github.com/lwlcom/cisco_exporter

FROM alpine:latest
WORKDIR /bin
COPY --from=builder /go/bin/cisco_exporter .
EXPOSE 9362
CMD ["sh"]