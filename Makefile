SHELL := /usr/bin/env bash

ROOT_DIR := $(CURDIR)
SERVER_DIR := $(ROOT_DIR)/src/server
UI_DIR := $(ROOT_DIR)/src/ui

.PHONY: swagger-backend build-backend dev-backend dev-backend-watch dev-frontend

swagger-backend:
	cd $(SERVER_DIR) && go generate ./...

build-backend: swagger-backend
	cd $(SERVER_DIR) && go build ./...

dev-backend: swagger-backend
	cd $(SERVER_DIR) && go run .

dev-backend-watch:
	cd $(SERVER_DIR) && air -c .air.toml

dev-frontend:
	cd $(UI_DIR) && elm make src/Main.elm --output=elm.js
