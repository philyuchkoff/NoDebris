.PHONY: build run test clean

build:
	go build -o nodebris ./cmd/nodebris/

run:
	go run ./cmd/nodebris/ serve

test:
	go test ./internal/... -v

clean:
	rm -f nodebris
