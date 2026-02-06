BINARY_NAME=api
MAIN_PATH=cmd/api/main.go

.PHONY: all build run test clean docker-up docker-down

all: build

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	go clean

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
