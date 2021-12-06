# https://stackoverflow.com/questions/714100/os-detecting-makefile
ifeq ($(OS),Windows_NT)
    CCFLAGS += -D WIN32
    ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
        CCFLAGS += -D AMD64
    else
        ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
            CCFLAGS += -D AMD64
        endif
        ifeq ($(PROCESSOR_ARCHITECTURE),x86)
            CCFLAGS += -D IA32
        endif
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        CCFLAGS += -D LINUX
    endif
    ifeq ($(UNAME_S),Darwin)
        CCFLAGS += -D OSX
    endif
    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
        CCFLAGS += -D AMD64
    endif
    ifneq ($(filter %86,$(UNAME_P)),)
        CCFLAGS += -D IA32
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
        CCFLAGS += -D ARM
    endif
endif

.PHONY: build
build:
	go build

.PHONY: build-static
build-static: pcap
ifeq ($(UNAME_S), Linux)
	LDFLAGS='-l./libpcap.a' CGO_ENABLED=1 \
	go build -ldflags '-linkmode external -extldflags -static' -o go-sniffer
endif

.PHONY: dep
dep: pcap

.PHONY: pcap
pcap:
ifeq ($(UNAME_S), Linux)
ifeq (,$(wildcard libpcap))
	git clone https://github.com/the-tcpdump-group/libpcap.git
	cd libpcap && ./configure && make
endif
endif

.PHONY: debug
debug:
	go build -gcflags "all=-N -l"
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./go-sniffer

.PHONY: clean
clean:
	rm -rf libpcap
	rm -f go-sniffer