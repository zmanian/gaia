all: get_vendor_deps test install

test:
	@go test `glide novendor`
	bash ./cmd/gaia/sh_tests/stake.sh
	
install:
	go install ./cmd/gaia

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install
