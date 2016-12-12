CWD=$(shell pwd)
export GOPATH:=$(CWD)/_vendor:$(CWD)

BIN=httpsocket_loader

.PHONY: clean depends linux bin/$(BIN)

all: bin/$(BIN)

bin/$(BIN):
	go install $(BIN)

linux:
	GOOS=linux go build $(BIN)

depends:
	go get github.com/gorilla/websocket


clean:
	rm -rf bin/
	rm -rf pkg/
	rm -rf _vendor/bin/
	rm -rf _vendor/pkg/
