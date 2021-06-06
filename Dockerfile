FROM golang as builder
ADD . /go/cisco_exporter/
WORKDIR /go/cisco_exporter
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/cisco_exporter

FROM alpine
ENV SSH_KEYFILE ""
ENV CONFIG_FILE "/config.yml"
ENV CMD_FLAGS ""
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /go/bin/cisco_exporter .
CMD ./cisco_exporter -ssh.keyfile=$SSH_KEYFILE -config.file=$CONFIG_FILE $CMD_FLAGS
EXPOSE 9362
