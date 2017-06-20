FROM ubuntu:latest
RUN apt-get clean
RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y iproute2 iputils-ping net-tools tcpdump dstat curl less vim sqlite3 jq git golang libpcre3-dev
RUN curl https://repo.powerdns.com/FD380FBB-pub.asc | apt-key add - 
RUN echo "deb [arch=amd64] http://repo.powerdns.com/ubuntu $(grep VERSION_CODENAME /etc/os-release | cut -f 2 -d '=')-auth-40 main" > /etc/apt/sources.list.d/pdns.list
RUN printf "Package: pdns-*\nPin: origin repo.powerdns.com\nPin-Priority: 600\n" > /etc/apt/preferences.d/pdns
RUN apt-get update 
RUN apt-get install -y pdns-server pdns-backend-sqlite3 pdns-tools
RUN mkdir -p /etc/powerdns
RUN sqlite3 /var/spool/powerdns/powerdns.db < /usr/share/doc/pdns-backend-sqlite3/schema.sqlite3.sql
RUN chown pdns:pdns /var/spool/powerdns/powerdns.db
RUN mkdir /tmp/gopath
ENV GOPATH /tmp/gopath
RUN go get -u github.com/alecthomas/gometalinter
ENV PATH ${PATH}:/tmp/gopath/bin
RUN cd /tmp && curl https://glide.sh/get | sh && rm -rf glide*
RUN mkdir -p /tmp/gopath/src/github.com/potix
RUN cd /tmp/gopath/src/github.com/potix && git clone https://github.com/potix/pdns-record-updater 
RUN cd /tmp/gopath/src/github.com/potix/pdns-record-updater && glide update && go build
ENTRYPOINT ["/tmp/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater"]
CMD ["-mode", "updater", "-config", "-config", "/etc/powerdns/pdns-record-updater.yaml"]
