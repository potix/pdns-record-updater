FROM potix/pdns-record-updater-base:0.1

RUN sigil -p -f /etc/powerdns/pdns-record-watcher.yaml.template > /etc/powerdns//etc/powerdns/pdns-record-watcher.yaml

EXPOSE 28001/tcp

ENTRYPOINT ["/root/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater"]
CMD ["-mode", "watcher", "-config", "/etc/powerdns/pdns-record-watcher.yaml"]
