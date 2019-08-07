FROM docker.io/ubuntu:bionic

RUN apt update && apt install --no-install-recommends -y software-properties-common

# Source: https://github.com/golang/go/wiki/Ubuntu
RUN add-apt-repository ppa:longsleep/golang-backports && \
    apt update &&                                        \
    apt install --no-install-recommends -y golang-go

# We are exiting zero because the installation of `resolvconf` fails.
# RUN apt-get update && \
RUN apt install --no-install-recommends -y \
      ca-certificates                      \
      gcc                                  \
      git                                  \
      libc-dev                             \
      sudo                                 \
      resolvconf;                          \
        exit 0

RUN rm -rf /var/lib/apt/lists/*

ENV GOPATH /root/go
RUN mkdir -p /root/go/bin
ENV PATH $PATH:/usr/local/go/bin:$GOPATH/bin
RUN go get -u github.com/onsi/ginkgo/ginkgo && \
    go get -u github.com/onsi/gomega/...

ADD mount-etc-resolv.conf.sh /
