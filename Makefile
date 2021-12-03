.PHONY: build
build:
	go build

.PHONY: build-static
build-static: pcap
	LDFLAGS='-l./libpcap.a' CGO_ENABLED=1 \
	go build -ldflags '-linkmode external -extldflags -static' -o go-sniffer

.PHONY: pcap
pcap:
ifeq (,$(wildcard libpcap))
	git clone https://github.com/the-tcpdump-group/libpcap.git
	cd libpcap && ./configure && make
else
	cd libpcap && make clean && ./configure && make
endif

.PHONY: debug
debug:
	go build -gcflags "all=-N -l"
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./go-sniffer