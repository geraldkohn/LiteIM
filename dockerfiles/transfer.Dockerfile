FROM golang as build

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /im
COPY . .

RUN chmod +x script/*.sh

RUN /bin/sh -c script/build_transfer.sh

FROM ubuntu:20.04

COPY --from=build /im/script /im/script
COPY --from=build /im/bin /im/bin

WORKDIR /im

CMD ["script/start_transfer.sh"]
