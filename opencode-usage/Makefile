.PHONY: build install clean test

BINARY=oc-usage

build:
	CGO_ENABLED=0 go build -o $(BINARY) .

install:
	CGO_ENABLED=0 go build -o $(GOPATH)/bin/$(BINARY) .

clean:
	rm -f $(BINARY)

test:
	go test ./...

run: build
	./$(BINARY)
