
EXECUTABLE_FILE_NAME = avatars

all: install build

install:
	go get -v ./...

run:
	go run *.go

build:
	go build -o $(EXECUTABLE_FILE_NAME)
	
clean:
	rm ./$(EXECUTABLE_FILE_NAME)

buildrun: build run

.PHONY: all run install build buildrun clean
