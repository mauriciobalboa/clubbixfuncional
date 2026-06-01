# syntax=docker/dockerfile:1.7

FROM golang:1.21-bookworm AS builder

WORKDIR /src/backend

ENV CGO_ENABLED=0 \
    GOOS=linux

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
COPY tools/ ./tools/

RUN go build -trimpath -ldflags="-s -w" -o /out/clubbix-server ./main.go

RUN go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./tools/healthcheck

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app

ENV PORT=8080 \
    ENVIRONMENT=production \
    DB_PATH=/app/clubbix.db

COPY --from=builder --chown=nonroot:nonroot /out/clubbix-server /app/clubbix-server
COPY --from=builder --chown=nonroot:nonroot /out/healthcheck /app/healthcheck
COPY --chown=nonroot:nonroot frontend /app/frontend

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD ["/app/healthcheck"]

USER nonroot:nonroot

ENTRYPOINT ["/app/clubbix-server"]
