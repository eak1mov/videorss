# https://github.com/yandex-cloud/fluent-bit-plugin-yandex/releases
FROM cr.yandex/yc/fluent-bit-plugin-yandex:v2.1.1-fluent-bit-2.1.7

# static busybox with wget for healthcheck
COPY --from=busybox:1.38.0-musl /bin/busybox /bin/busybox

COPY fluentbit.conf /fluent-bit/etc/fluent-bit.conf
COPY parsers.conf /fluent-bit/etc/parsers.conf
