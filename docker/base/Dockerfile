FROM ubuntu:latest
RUN apt-get clean
RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y iproute2 iputils-ping net-tools tcpdump dstat curl less vim sqlite3 jq git golang libpcre3-dev letsencrypt
RUN curl https://repo.powerdns.com/FD380FBB-pub.asc | apt-key add - 
RUN echo "deb [arch=amd64] http://repo.powerdns.com/ubuntu $(grep VERSION_CODENAME /etc/os-release | cut -f 2 -d '=')-auth-40 main" > /etc/apt/sources.list.d/pdns.list
RUN printf "Package: pdns-*\nPin: origin repo.powerdns.com\nPin-Priority: 600\n" > /etc/apt/preferences.d/pdns
RUN apt-get update 
RUN apt-get install -y pdns-server pdns-backend-sqlite3 pdns-tools
RUN mkdir -p /etc/powerdns/static
RUN sqlite3 /var/spool/powerdns/powerdns.db < /usr/share/doc/pdns-backend-sqlite3/schema.sqlite3.sql
RUN chown pdns:pdns /var/spool/powerdns/powerdns.db
RUN mkdir /root/gopath
ENV GOPATH /root/gopath
RUN go get -u github.com/alecthomas/gometalinter
ENV PATH ${PATH}:/root/gopath/bin
RUN cd /root && curl https://glide.sh/get | sh && rm -rf glide*
RUN mkdir -p /root/gopath/src/github.com/potix
RUN cd /root/gopath/src/github.com/potix && git clone https://github.com/potix/pdns-record-updater 
RUN cd /root/gopath/src/github.com/potix/pdns-record-updater && glide update && go build
RUN chmod 755 /root/gopath/src/github.com/potix/pdns-record-updater/dockerfile/updater/run.sh
RUN curl -L https://github.com/gliderlabs/sigil/releases/download/v0.4.0/sigil_0.4.0_Linux_x86_64.tgz | tar -zxvC /usr/local/bin
RUN openssl req -x509 -sha256 -nodes -days 3650 -newkey rsa:2048 -subj /CN=localhost -keyout /etc/powerdns/pdns-record-updater-api-server.key -out /etc/powerdns/pdns-record-updater-api-server.cert
# environment
# ${PDNS_API_KEY:-pdns_api_key}
# ${PDNS_PACKET_CACHE_TTL:-5}
# ${PDNS_NEGATIVE_QUERY_CACHE_TTL:-60}
# ${PDNS_QUERY_CACHE_TTL:-5}
# ${PDNS_SOA_MINIMUM_TTL:-60}
COPY pdns.conf.template /etc/powerdns/pdns.conf.template
# environment
# ${WATCHER_NOTIFY_SUBJECT:-}
# ${WATCHER_NOTIFY_BODY:-}
# ${WATCHER_NOTIFY_SMTP_SERVER_PORT:-}
# ${WATCHER_NOTIFY_SMTP_TO:-}
# ${WATCHER_NOTIFY_SMTP_FROM:-}
# ${WATCHER_NOTIFY_SMTP_USERNAME:-}
# ${WATCHER_NOTIFY_SMTP_PASSWORD:-}
# ${WATCHER_NOTIFY_SMTP_AUTH_TYPE:-plain}
# ${WATCHER_NOTIFY_SMTP_USE_STARTTLS:-false}
# ${WATCHER_NOTIFY_SMTP_USE_TLS:-false}
# ${WATCHER_NOTIFY_SMTP_TLS_SKIP_VERIFY:-false}
# ${WATCHER_API_SERVER_USE_TLS:-false}
# ${WATCHER_API_SERVER_USERNAME:-}
# ${WATCHER_API_SERVER_PASSWORD:-}
COPY pdns-record-updater-updater.yaml.template /etc/powerdns/pdns-record-updater-updater.yaml.template
# environment
# ${WATCHER_API_SERVER_URL_LIST:?}
# ${WATCHER_TLS_SKIP_VERIFY:-false}
# ${WATCHER_API_SERVER_USERNAME:-}
# ${WATCHER_API_SERVER_PASSWORD:-}
# ${PDNS_API_KEY:-pdns_api_key}
COPY pdns-record-updater-watcher.yaml.template /etc/powerdns/pdns-record-updater-watcher.yaml.template
COPY run_updater.sh /usr/sbin/run_updater.sh
# other environment
# WATCHER_API_SERVER_CERT
# WATCHER_API_SERVER_PRIVATE_KEY
