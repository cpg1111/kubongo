ifeq ($(OS),Windows_NT)
	LDD_CMD = echo "Pff. Seriously? Windows?"; exit 1;
    CCFLAGS += -D WIN32
    ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
        CCFLAGS += -D AMD64
    endif
    ifeq ($(PROCESSOR_ARCHITECTURE),x86)
        CCFLAGS += -D IA32
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        CCFLAGS += -D LINUX
		LDD_CMD = ldd
    endif
    ifeq ($(UNAME_S),Darwin)
        CCFLAGS += -D OSX
		LDD_CMD = otool -l
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
all: build
get-deps:
	go get ./...
	godep restore ./...
build:
	godep restore ./...
	go build -o ./kubongo github.com/cpg1111/kubongo/
	$(LDD_CMD) ./kubongo
	file ./kubongo
install:
	mv ./kubongo /usr/bin/kubongo
clean:
	rm -rf $GOPATH/bin/github.com/cpg1111/kubongo $GOPATH/pkg/github.com/cpg1111/kubongo $GOPATH/src/github.com/cpg1111/kubongo/kubongo
test:
	go test -v ./...
uninstall:
	rm -rf /usr/bin/kubongo
