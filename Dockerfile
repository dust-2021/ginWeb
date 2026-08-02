FROM alpine:latest

VOLUME /var/log /var/log

EXPOSE 8000/tcp
EXPOSE 8001/udp
EXPOSE 8003/udp

USER root

COPY config-docker.yaml ./config.yaml
COPY exampleApp ./app


CMD ./app