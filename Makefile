all: get_vendor_deps install test

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install
	
install:
	go install ./cmd/gaia

test2: test_unit test_cli

test:
	@go test `glide novendor`

test_cli:
	bash ./cmd/gaia/sh_tests/stake.sh

