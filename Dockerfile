FROM golang:1.12.4 as builder
WORKDIR /go/src/github.com/videocoin/cloud-miners
COPY . .
RUN make build

FROM bitnami/minideb:jessie
COPY --from=builder /go/src/github.com/videocoin/cloud-miners/bin/miners /opt/videocoin/bin/miners
CMD ["/opt/videocoin/bin/miners"]