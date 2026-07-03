.PHONY: build test lint bench clean

build:
	go build -o grep-tool .

test:
	go test ./...

lint:
	golangci-lint run ./...

bench:
	go test -bench=. ./...

clean:
	rm -f grep-tool
