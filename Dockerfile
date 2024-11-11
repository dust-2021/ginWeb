FROM ubuntu:latest
#FROM mysql:5.7
#FROM redis:latest
#
#ENV MYSQL_DATABASE "gin_web"

COPY product.yaml ./config.yaml
COPY app .


CMD ./app