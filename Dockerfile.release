FROM alpine:3.22 AS runtime

ARG TARGETARCH

RUN addgroup -S -g 10001 rewind \
    && adduser -S -D -H -u 10001 -G rewind rewind \
    && mkdir -p /data \
    && chown 10001:10001 /data

COPY --from=release-binaries /rewind-linux-${TARGETARCH} /usr/local/bin/rewind

ENV REWIND_LISTEN=0.0.0.0:8080

LABEL org.opencontainers.image.source="https://github.com/YangtzeCoder/rewind"

USER 10001:10001
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/rewind"]
CMD ["serve", "--data-dir", "/data"]
