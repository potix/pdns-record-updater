FROM potix/pdns-record-updater-base:0.1

RUN sigil -p -f /etc/powerdns/pdns.conf.template > /etc/powerdns/pdns.conf && \
    sigil -p -f /etc/powerdns/pdns-record-updater.yaml.template > /etc/powerdns//etc/powerdns/pdns-record-updater.yaml

EXPOSE 53/udp 53/tcp

ENTRYPOINT ["/usr/sbin/run_updater.sh"]
CMD ["-mode", "updater", "-config", "/etc/powerdns/pdns-record-updater.yaml"]
