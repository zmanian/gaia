all: get_vendor_deps install test

get_vendor_deps:
	go get github.com/golang/dep/cmd/dep
	dep ensure
	
install:
	go install ./cmd/gaia

test:
	@go test $(shell go list ./... | grep -v vendor)

test_cli:
	bash ./cmd/gaia/sh_tests/stake.sh
