FROM ubuntu:18.04
RUN apt-get update


# install golang
RUN apt-get install -qy wget git gcc graphviz
RUN GO_VERSION=go1.14.linux-amd64.tar.gz &&\
    cd /opt &&\
    wget --quiet https://dl.google.com/go/$GO_VERSION &&\
    tar -C /usr/local -xzf $GO_VERSION &&\
    rm /opt/$GO_VERSION

RUN mkdir -p /go/src
ENV GOPATH=/go
ENV GOROOT=/usr/local/go
ENV PATH="$PATH:$GOROOT/bin:$GOPATH/bin"

RUN TMP='package main; func main() {println("Hello world")}' &&\
    mkdir -p /go/src/hello &&\
    echo $TMP> /go/src/hello/hello.go


# optional packages: vim 54Mb, net 2Mb
RUN apt-get install -qy vim
RUN apt-get install -qy net-tools telnet
