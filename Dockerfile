FROM golang:1.23.1
#
ENV MYSQL_DATABASE "gin_web"
VOLUME /var/log /var/log

RUN go build -o app .

COPY product.yaml ./config.yaml
COPY app .


CMD ./app