#!/bin/sh

running="true"

finish() {
	echo "stop pdns_server"
	kill $(echo $(ps ax | grep pdns_server | grep -v pdns_server-instance | grep -v grep) | awk '{ print $1 }') 
	echo "stop pdns-record-updater"
	kill $(echo $(ps ax | grep pdns-record-updater | grep -v grep) | awk '{ print $1 }') 
	running="false"
}

trap finish INT QUIT TERM

echo "start $0"


echo "start pdns-record-updater"
/root/gopath/src/github.com/potix/pdns-record-updater/pdns-record-updater $@

sleep 10 # WORKAROUND: can not detect initialize finished

echo "start pdns_server"
/usr/sbin/pdns_server

while true
do
	if [ "${running}" = "false" ]; then
		echo "stop $0"
		exit
	fi
	sleep 1
done
