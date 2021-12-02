.PHONY: build
build:
	go build

.PHONY: build-static
build-static: pcap
	LDFLAGS='-l./libpcap.a' CGO_ENABLED=1 \
	go build -ldflags '-linkmode external -extldflags -static' -o go-sniffer

.PHONY: pcap
pcap:
	@if [ -a libpcap ]; \
	then \
		cd libpcap && git pull ; \
		if [ -a libpcap.a ]; \
		then \
			echo "libpcap has been builded"; \
		else \
			./configure && make ; \
		fi \
	else \
		git clone https://github.com/the-tcpdump-group/libpcap.git ; \
		cd libpcap; \
		if [ -a libpcap.a ]; \
		then \
			echo "libpcap has been builded"; \
		else \
			./configure && make ; \
		fi \
	fi
