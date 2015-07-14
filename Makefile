
EXECUTABLE_FILE_NAME = avatars

all: install build

install:
	go get -v ./...

build:
	go build -o $(EXECUTABLE_FILE_NAME)

run: build
	./$(EXECUTABLE_FILE_NAME)

test:
	go test -cover -short ./...

clean:
	rm ./$(EXECUTABLE_FILE_NAME)

.PHONY: all install build run test clean
