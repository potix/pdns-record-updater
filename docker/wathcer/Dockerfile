FROM ubuntu:latest
RUN apt-get clean
RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y iproute2 iputils-ping net-tools tcpdump dstat curl less vim letsencrypt jq git golang libpcre3-dev
RUN mkdir -p /etc/powerdns
RUN mkdir /root/gopath
ENV GOPATH /root/gopath
RUN go get -u github.com/alecthomas/gometalinter
ENV PATH ${PATH}:/root/gopath/bin
RUN cd /root && curl https://glide.sh/get | sh && rm -rf glide*
RUN mkdir -p /root/gopath/src/github.com/potix
RUN cd /root/gopath/src/github.com/potix && git clone https://github.com/potix/pdns-record-updater 
RUN cd /root/gopath/src/github.com/potix/pdns-record-updater && glide update && go build
RUN curl -L https://github.com/gliderlabs/sigil/releases/download/v0.4.0/sigil_0.4.0_Linux_x86_64.tgz | tar -zxvC /usr/local/bin


EXPOSE 28001/tcp


ENTRYPOINT ["/root/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater"]
CMD ["-mode", "watcher", "-config", "-config", "/etc/powerdns/pdns-record-updater.yaml"]
