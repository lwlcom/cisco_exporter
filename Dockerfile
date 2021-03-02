FROM golang:alpine as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPATH=/build

WORKDIR /build

COPY . .


RUN apk add git && go get -d -v  || true
RUN go build -o cisco_exporter .

FROM scratch

COPY --from=builder /build/cisco_exporter /

EXPOSE 9362

ENTRYPOINT ["/cisco_exporter"]
