# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && \
    adduser -D -H -u 10001 appuser && \
    mkdir -p /data/uploads && chown appuser /data/uploads
COPY --from=build /out/api /api
ENV UPLOAD_DIR=/data/uploads
USER appuser
EXPOSE 8080
ENTRYPOINT ["/api"]