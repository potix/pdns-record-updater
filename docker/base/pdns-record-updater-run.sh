#!/bin/sh

running="true"

finish() {
	if [ "${PDNS_RECORD_UPDATER_MODE}" == "updater" ]; then
		echo "stop pdns_server"
		kill $(echo $(ps ax | grep pdns_server | grep -v pdns_server-instance | grep -v grep) | awk '{ print $1 }') 
		echo "stop pdns-record-updater"
		kill $(echo $(ps ax | grep pdns-record-updater | grep "-mode updater" | grep -v grep) | awk '{ print $1 }') 
	elif [ "${PDNS_RECORD_UPDATER_MODE}" == "watcher" ]; then
		echo "stop pdns-record-updater"
		kill $(echo $(ps ax | grep pdns-record-updater | grep "-mode watcher" | grep -v grep) | awk '{ print $1 }') 
	elif [ "${PDNS_RECORD_UPDATER_MODE}" == "manager" ]; then
		echo "stop pdns-record-updater"
		kill $(echo $(ps ax | grep pdns-record-updater | grep "-mode manager" | grep -v grep) | awk '{ print $1 }') 
	fi
	running="false"
}

start_updater() {
	echo "create config"
	sigil -p -f /etc/powerdns/pdns.conf.template > /etc/powerdns/pdns.conf
	sigil -p -f /etc/powerdns/pdns-record-updater-updater.yaml.template > /etc/powerdns/pdns-record-updater.yaml
	echo "init database"
	rm -f /var/spool/powerdns/powerdns.db.initialized
	rm -f /var/spool/powerdns/powerdns.db
	sqlite3 /var/spool/powerdns/powerdns.db < /usr/share/doc/pdns-backend-sqlite3/schema.sqlite3.sql
	chown pdns:pdns /var/spool/powerdns/powerdns.db
	echo "start pdns-record-updater"
	/root/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater -mode ${PDNS_RECORD_UPDATER_MODE} -config /etc/powerdns/pdns-record-updater.yaml &
	echo "wait initialize"
	while true 
	do
		sleep 1
		if [ "${running}" = "false" ]; then
			sleep 1
			echo "stop $0"
			exit
		fi
		if [ -f /var/spool/powerdns/powerdns.db.initialized ]; then
			break
		fi
	done
	echo "start pdns_server"
	/usr/sbin/pdns_server
}

start_watcher() {
	echo "create config"
	sigil -p -f /etc/powerdns/pdns-record-updater-watcher.yaml.template > /etc/powerdns/pdns-record-updater.yaml
	echo "start pdns-record-updater"
	/root/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater -mode ${PDNS_RECORD_UPDATER_MODE} -config /etc/powerdns/pdns-record-updater.yaml &
}

start_manager() {
	echo "create config"
	sigil -p -f /etc/powerdns/pdns-record-updater-manager.yaml.template > /etc/powerdns/pdns-record-updater.yaml
	echo "start pdns-record-updater"
	/root/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater -mode ${PDNS_RECORD_UPDATER_MODE} -config /etc/powerdns/pdns-record-updater.yaml &
}

trap finish INT QUIT TERM

echo "start $0"

if [ "${PDNS_RECORD_UPDATER_MODE}" == "updater" ]; then
	start_updater
elif [ "${PDNS_RECORD_UPDATER_MODE}" == "watcher" ]; then
	start_watcher
elif [ "${PDNS_RECORD_UPDATER_MODE}" == "manager" ]; then
	start_manager
fi

while true
do
	if [ "${running}" = "false" ]; then
		sleep 1
		echo "stop $0"
		exit
	fi
	sleep 1
done
