#!/bin/sh

running="true"

finish() {
	echo "stop pdns_server"
	kill $(echo $(ps ax | grep pdns_server | grep -v pdns_server-instance | grep -v grep) | awk '{ print $1 }') 
	echo "stop pdns-record-updater"
	kill $(echo $(ps ax | grep pdns-record-updater | grep -v watcher | grep -v grep) | awk '{ print $1 }') 
	running="false"
}

trap finish INT QUIT TERM

echo "start $0"

echo "init database"
rm -f /var/spool/powerdns/powerdns.db.initialized
rm -f /var/spool/powerdns/powerdns.db
sqlite3 /var/spool/powerdns/powerdns.db < /usr/share/doc/pdns-backend-sqlite3/schema.sqlite3.sql
chown pdns:pdns /var/spool/powerdns/powerdns.db

echo "start pdns-record-updater"
/root/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater $@ &

while true 
do
	if [ "${running}" = "false" ]; then
		sleep 1
		echo "stop $0"
		exit
	fi
	sleep 1
	if [ -f /var/spool/powerdns/powerdns.db.initialized ]; then
		break
	fi
done

echo "start pdns_server"
/usr/sbin/pdns_server

while true
do
	if [ "${running}" = "false" ]; then
		sleep 1
		echo "stop $0"
		exit
	fi
	sleep 1
done
