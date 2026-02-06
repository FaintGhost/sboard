FROM node:22-alpine AS web

WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web ./
RUN npm run build

FROM golang:1.25 AS builder

WORKDIR /src
COPY go.mod ./

# The repo does not necessarily ship with go.sum (fresh clone / early stage),
# so we must allow the image build to generate a complete go.sum on the fly.
# Some environments also set GOFLAGS=-mod=readonly; use -mod=mod explicitly.
RUN go mod download
COPY . .

# `go mod tidy` does not accept `-mod=...`. Also avoid inheriting GOFLAGS=-mod=readonly.
RUN GOFLAGS= go mod tidy
RUN CGO_ENABLED=0 go build -mod=mod -trimpath -ldflags "-s -w" -o /out/sboard-panel ./cmd/panel

FROM debian:stable-slim

WORKDIR /app
COPY --from=builder /out/sboard-panel /app/sboard-panel
COPY --from=web /web/dist /app/web/dist

ENV PANEL_HTTP_ADDR=:8080
ENV PANEL_DB_PATH=/data/panel.db
ENV PANEL_SERVE_WEB=true
ENV PANEL_WEB_DIR=/app/web/dist

EXPOSE 8080
ENTRYPOINT ["/app/sboard-panel"]
