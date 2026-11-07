FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine AS backend

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /src/frontend/dist ./internal/server/web/dist
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /out/rewind ./cmd/rewind

FROM alpine:3.22 AS runtime

RUN addgroup -S -g 10001 rewind \
    && adduser -S -D -H -u 10001 -G rewind rewind \
    && mkdir -p /data \
    && chown 10001:10001 /data

COPY --from=backend /out/rewind /usr/local/bin/rewind

ENV REWIND_LISTEN=0.0.0.0:8080

LABEL org.opencontainers.image.source="https://github.com/YangtzeCoder/rewind"

USER 10001:10001
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/rewind"]
CMD ["serve", "--data-dir", "/data"]
