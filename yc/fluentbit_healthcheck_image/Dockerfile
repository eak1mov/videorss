FROM alpine:latest
RUN apk add --no-cache curl jq
CMD ["curl", "--fail", "--silent", "--output", "/dev/null", "--retry", "2", "--retry-connrefused", "http://localhost:2020/api/v1/health"]
