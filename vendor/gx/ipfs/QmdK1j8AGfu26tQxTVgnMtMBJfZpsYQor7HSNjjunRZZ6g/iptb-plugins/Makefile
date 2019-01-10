IPTB_ROOT ?=$(HOME)/testbed

all: iptb

deps:
	gx install

iptb: deps
	gx-go rw
	(cd iptb; go build)
	gx-go uw
CLEAN += iptb/iptb

ipfslocal: deps
	gx-go rw
	(cd local/plugin; go build -buildmode=plugin -o ../../build/localipfs.so)
	gx-go uw
CLEAN += build/localipfs.so

p2pdlocal: deps
	gx-go rw
	(cd localp2pd/plugin; go build -buildmode=plugin -o ../../build/localp2pd.so)
	gx-go uw
CLEAN += build/localp2pd.so

ipfsdocker: deps
	gx-go rw
	(cd docker/plugin; go build -buildmode=plugin -o ../../build/dockeripfs.so)
	gx-go uw
CLEAN += build/dockeripfs.so

ipfsbrowser:
	gx-go rw
	(cd browser/plugin; go build -buildmode=plugin -o ../../build/browseripfs.so)
	gx-go uw
CLEAN += build/browseripfs.so

install: deps
	gx-go rw
	(cd iptb; go install)
	gx-go uw

clean:
	rm ${CLEAN}

.PHONY: all clean ipfslocal p2pdlocal ipfsdocker ipfsbrowser
