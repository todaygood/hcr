

# build: Build the  binary
.PHONY: build
build:
	@CGO_ENABLED=0 go build --tags=release -a -o hcr ./cmd/main.go  ./cmd/appcontext.go
