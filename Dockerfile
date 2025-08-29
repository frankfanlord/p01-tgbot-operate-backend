FROM scratch

WORKDIR /app

COPY ./operate .
COPY ./config ./config
COPY ./localtime /etc/localtime

ENTRYPOINT ["./operate"]
