FROM golang as build

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /im
COPY . .

RUN chmod +x script/*.sh

RUN /bin/sh -c script/build_gateway.sh

FROM ubuntu:20.04

COPY --from=build /im/script /im/script
COPY --from=build /im/bin /im/bin
COPY --from=build /im/config /im/config

WORKDIR /im

CMD ["script/start_gateway.sh"]
