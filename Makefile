WEB_DIR := web
BIN_DIR := bin
BIN := $(BIN_DIR)/tarc

.PHONY: all build web-build backend-build run clean

all: build

build: web-build backend-build

web-build:
	cd $(WEB_DIR) && npm ci && npm run build

backend-build: web-build
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) .

run: build
	./$(BIN)

clean:
	rm -rf $(BIN_DIR) $(WEB_DIR)/dist