FROM golang:1.14.4-alpine

WORKDIR /src
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -o iptv-proxy ./cmd/iptv-proxy/main.go

FROM scratch
COPY --from=0  /src/iptv-proxy /
ENTRYPOINT ["/iptv-proxy"]
