APP_NAME=web_ws
IMAGE_NAME=web_ws
TAG?=latest

.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/server

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_NAME):$(TAG) .

.PHONY: docker-run
docker-run:
	docker run --rm -p 8080:8080 -e PORT=8080 -e JWT_SECRET=CHANGE_ME_IN_PROD $(IMAGE_NAME):$(TAG)

.PHONY: compose-up
compose-up:
	docker compose up -d --build

.PHONY: compose-down
compose-down:
	docker compose down

