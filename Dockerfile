FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install --quiet --yes net-tools wget tree htop vim telnet

# install golang
RUN mkdir -p /home/tungdt/go/src
RUN cd /opt && \
    wget --quiet https://dl.google.com/go/go1.12.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.12.linux-amd64.tar.gz
ENV PATH $PATH:/usr/local/go/bin
ENV GOPATH /home/tungdt/go
RUN apt-get install -qy git gcc graphviz

# get dependencies
COPY ./go.mod /home/tungdt
COPY ./go.sum /home/tungdt
RUN go mod download
ENV GO111MODULE=on
RUN mkdir -p /home/tungdt/go/src/github.com/daominah
COPY . /home/tungdt/go/src/github.com/daominah/gomicrokit
WORKDIR /home/tungdt/go/src/github.com/daominah/gomicrokit
