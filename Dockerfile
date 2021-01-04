ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL org.opencontainers.image.source https://github.com/swoga/alertmanager-filter

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/alertmanager-filter /bin/alertmanager-filter
COPY example.yml /etc/alertmanager-filter/config.yml

RUN chown -R nobody:nobody /etc/alertmanager-filter

USER nobody
EXPOSE 80

ENTRYPOINT ["/bin/alertmanager-filter"]
CMD ["--config.file=/etc/alertmanager-filter/config.yml"]