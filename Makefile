.PHONY: build run test clean

build:
	go build -o nodebris ./cmd/bookmark-cleaner/

run:
	go run ./cmd/bookmark-cleaner/ serve

test:
	go test ./internal/... -v

clean:
	rm -f nodebris
