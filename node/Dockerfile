FROM golang:1.25 AS builder

WORKDIR /src
COPY go.mod ./

# go.sum may be generated in this layer if missing.
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/sboard-node ./cmd/node

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /out/sboard-node /app/sboard-node

ENV NODE_HTTP_ADDR=:3000
ENV NODE_LOG_LEVEL=info

EXPOSE 3000
ENTRYPOINT ["/app/sboard-node"]

