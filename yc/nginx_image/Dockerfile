# FROM nginx:stable-alpine
# FROM nginx:stable-alpine-slim
FROM nginxinc/nginx-unprivileged:stable-alpine-slim
COPY nginx.conf /etc/nginx/conf.d/default.conf
ENV NGINX_ENTRYPOINT_QUIET_LOGS=1
