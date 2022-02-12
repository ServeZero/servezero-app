# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=main
BUILD_TAG=release
BUILD_DIR_NAME=bin
INSTALL_DATABASE=install/database.sh

all: run

build:
	$(GOBUILD) -o ./$(BUILD_DIR_NAME)/$(BINARY_NAME) -tags=$(BUILD_TAG) -v ./

test:
	$(GOTEST) -v ./

clean:
	$(GOCLEAN)
	rm -rf ./$(BUILD_DIR_NAME)/$(BINARY_NAME)
	rm -rf ./$(BUILD_DIR_NAME)

run: export GIN_MODE=release
run:
	$(GOBUILD) -o ./$(BUILD_DIR_NAME)/$(BINARY_NAME) -tags=$(BUILD_TAG) -v ./
	./$(BUILD_DIR_NAME)/$(BINARY_NAME)

restart:
	kill -INT $$(cat pid)
	$(GOBUILD) -o ./$(BUILD_DIR_NAME)/$(BINARY_NAME) -tags=$(BUILD_TAG) -v ./
	./$(BUILD_DIR_NAME)/$(BINARY_NAME)

db:
	./$(INSTALL_DATABASE)
