all:
	glide update
	cd manager && go-bindata -pkg manager asset/...
	go build
clea:
	rm -f pdns-record-updater
