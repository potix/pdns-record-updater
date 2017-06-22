all:
	glide update
	cd manager && go-bindata -pkg manager asset/...
	go build
clean:
	rm -f pdns-record-updater
