GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)


compile: 
	@-$(MAKE) get
	@-$(MAKE) build

build:
	@echo "  >  Building snapshot-archive..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o $(GOBIN)/snapshot-archive $(GOFILES)

install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

clean:
	@-rm $(GOBIN)/*
	@-rm vendor/* -rf
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

get:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get -u github.com/ledgerwatch/turbo-geth@58c3371dbcb1bdd4164a587e33249d1c49c30036

test:
	@GOPATH=$(GOPATH) go test