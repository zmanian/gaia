all: get_vendor_deps install test

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install
	
install:
	go install ./cmd/gaia

test:
	@go test `glide novendor`

test_cli: install
	go run test/test.go all
