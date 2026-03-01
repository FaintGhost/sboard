FROM golang:1.25 AS builder

WORKDIR /src
COPY sb-client-entrypoint.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/sb-client-entrypoint ./sb-client-entrypoint.go

FROM gzxhwq/sing-box:1.12.19

COPY --from=builder /out/sb-client-entrypoint /usr/local/bin/sb-client-entrypoint

ENTRYPOINT ["/usr/local/bin/sb-client-entrypoint"]
CMD ["run", "-D", "/etc/sing-box", "-C", "/etc/sing-box"]
