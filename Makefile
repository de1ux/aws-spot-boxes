PACKAGE = github.com/de1ux/aws-spot-boxes
GOPATH = $(CURDIR)/.gopath
GOBIN = $(CURDIR)/.gopath/bin
PATH := $(GOBIN):$(PATH)
GO = go

api:
	go get -u github.com/golang/protobuf/protoc-gen-go
	rm -rf $(CURDIR)/app/src/generated generated
	mkdir -p $(CURDIR)/app/src/generated generated
	protoc \
	    --go_out=plugins=grpc:"generated" \
	    api/*

.PHONY: api
