FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

RUN adduser -D -g '' appuser

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o app cmd/api/main.go

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /build/migrations /migrations

COPY --from=builder /build/app /app

USER appuser

EXPOSE 8080

ENV GIN_MODE=release
ENV TZ=UTC

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app", "health"] || exit 1

ENTRYPOINT ["/app"]