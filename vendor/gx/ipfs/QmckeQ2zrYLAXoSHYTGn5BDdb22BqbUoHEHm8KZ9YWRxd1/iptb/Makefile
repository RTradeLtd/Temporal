CLEAN =

all: iptb

iptb:
	go build
CLEAN += iptb

install:
	go install

test:
	make -C sharness all

clean:
	rm $(CLEAN)

.PHONY: all test iptb install plugins clean
