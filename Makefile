SHELL := /usr/bin/env bash

ROOT_DIR := $(CURDIR)
SERVER_DIR := $(ROOT_DIR)/src/server
UI_DIR := $(ROOT_DIR)/src/ui
BACKEND_PROXY_URL := http://localhost:8081/api
UI_BUILD_DIR := $(UI_DIR)/out
UI_BUILD_OUTPUT := out/app.js
UI_START_PAGE := index.html
UI_STATIC_FILES := index.html favicon.ico

.PHONY: backend backend-run backend-watch backend-swagger frontend frontend-watch _frontend-build-dir _frontend-copy-static

backend-swagger:
	cd $(SERVER_DIR) && go generate ./...

backend: backend-swagger
	cd $(SERVER_DIR) && go build ./...

backend-run: backend-swagger
	cd $(SERVER_DIR) && go run .

backend-watch:
	cd $(SERVER_DIR) && air -c .air.toml

_frontend-build-dir:
	mkdir -p $(UI_BUILD_DIR)

_frontend-copy-static: _frontend-build-dir
	cp $(addprefix $(UI_DIR)/,$(UI_STATIC_FILES)) $(UI_BUILD_DIR)/

frontend: _frontend-copy-static
	cd $(UI_DIR) && elm make src/Main.elm --output=$(UI_BUILD_OUTPUT)

frontend-watch: _frontend-copy-static
	cd $(UI_DIR) && elm-live src/Main.elm --port=8082 --pushstate --dir=out --start-page=$(UI_START_PAGE) --proxy-host=$(BACKEND_PROXY_URL) --proxy-prefix=/api --open -- --output=$(UI_BUILD_OUTPUT)
