BINARY_NAME=ask

all: test build

build:
	go build -o $(BINARY_NAME) main.go

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

run:
	go run main.go

deps:
	go mod download

fmt:
	go fmt ./...

vet:
	go vet ./...

install:
	go install
