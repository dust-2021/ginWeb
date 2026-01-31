FROM golang:1.25.1
#
ENV MYSQL_DATABASE "gin_web"
VOLUME /var/log /var/log

RUN go build -o app .

EXPOSE 8000/tcp
EXPOSE 8001/udp
EXPOSE 8003/udp

USER root

COPY product.yaml ./config.yaml
COPY app .


CMD ./app