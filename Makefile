.DEFAULT_GOAL := build
.PHONY: build clean push

IMAGE_NAME := libli/ipsync
DOCKER_REPO := docker.io
VERSION := 1.0

build:
	docker build --platform linux/amd64 -t $(DOCKER_REPO)/$(IMAGE_NAME):latest -t $(DOCKER_REPO)/$(IMAGE_NAME):$(VERSION) --progress=plain .
push:
	@echo "Pushing to docker hub"
	docker login -u libli -p $(DOCKER_PASSWORD) $(DOCKER_REPO)
	docker push $(DOCKER_REPO)/$(IMAGE_NAME) -a
clean:
	docker rmi $(docker images -q $(DOCKER_REPO)/$(IMAGE_NAME))