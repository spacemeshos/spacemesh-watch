FROM golang:1.16-alpine as builder

ADD . /spacemesh-watch
RUN cd /spacemesh-watch && go build

FROM alpine:latest
COPY --from=builder /spacemesh-watch/spacemesh-watch /usr/local/bin/

ENTRYPOINT ["spacemesh-watch"]
