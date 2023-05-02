FROM golang as build

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /im
COPY ../. .

WORKDIR /im/script
RUN chomd +x *.sh

RUN /bin/sh -c ./build_transfer.sh

FROM ubuntu:20.04

COPY --from=build /im/script /im/script
COPY --from=build /im/bin /im/bin

WORKDIR /im/script

CMD ["./start_transfer.sh"]
