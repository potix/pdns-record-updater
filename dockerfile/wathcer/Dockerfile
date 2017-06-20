FROM ubuntu:latest
RUN apt-get clean
RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y iproute2 iputils-ping net-tools tcpdump dstat curl less vim letsencrypt jq git golang libpcre3-dev
RUN mkdir -p /etc/powerdns
RUN mkdir /tmp/gopath
ENV GOPATH /tmp/gopath
RUN go get -u github.com/alecthomas/gometalinter
ENV PATH ${PATH}:/tmp/gopath/bin
RUN cd /tmp && curl https://glide.sh/get | sh && rm -rf glide*
RUN mkdir -p /tmp/gopath/src/github.com/potix
RUN cd /tmp/gopath/src/github.com/potix && git clone https://github.com/potix/pdns-record-updater 
RUN cd /tmp/gopath/src/github.com/potix/pdns-record-updater && glide update && go build
ENTRYPOINT ["/tmp/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater"]
CMD ["-mode", "watcher", "-config", "-config", "/etc/powerdns/pdns-record-updater.yaml"]
