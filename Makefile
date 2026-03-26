SHELL := /usr/bin/env bash

ROOT_DIR := $(CURDIR)
SERVER_DIR := $(ROOT_DIR)/src/server
UI_DIR := $(ROOT_DIR)/src/ui
BACKEND_PROXY_URL := http://localhost:8081/api
UI_BUILD_DIR := $(UI_DIR)/out
UI_ELM_OUTPUT := out/elm.js
UI_START_PAGE := index.html
UI_STATIC_FILES := index.html favicon.ico style.css
UI_BOOTSTRAP_FILE := src/app.js
UI_INTEGRATIONS_DIR := src/integrations

.PHONY: backend backend-run backend-watch backend-swagger frontend frontend-watch _frontend-build-dir _frontend-copy-static _frontend-copy-bootstrap _frontend-copy-integrations

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

_frontend-copy-bootstrap: _frontend-build-dir
	cp $(UI_DIR)/$(UI_BOOTSTRAP_FILE) $(UI_BUILD_DIR)/app.js

_frontend-copy-integrations: _frontend-build-dir
	rm -rf $(UI_BUILD_DIR)/integrations
	cp -R $(UI_DIR)/$(UI_INTEGRATIONS_DIR) $(UI_BUILD_DIR)/integrations

frontend: _frontend-copy-static _frontend-copy-bootstrap _frontend-copy-integrations
	cd $(UI_DIR) && elm make src/Main.elm --output=$(UI_ELM_OUTPUT)

frontend-watch: _frontend-copy-static _frontend-copy-bootstrap _frontend-copy-integrations
	cd $(UI_DIR) && elm-live src/Main.elm --port=8082 --pushstate --dir=out --start-page=$(UI_START_PAGE) --proxy-host=$(BACKEND_PROXY_URL) --proxy-prefix=/api --open -- --output=$(UI_ELM_OUTPUT)
