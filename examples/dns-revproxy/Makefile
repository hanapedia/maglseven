# Variables
DOCKER_HUB_USERNAME := hexagonbenchmark
REPOSITORY_NAME := maglseven
TAG := dns-revproxy
IMAGE_NAME := $(DOCKER_HUB_USERNAME)/$(REPOSITORY_NAME):$(TAG)

	BUILD_COMMAND=docker buildx build --platform linux/amd64 -t $(IMAGE_NAME) -f Dockerfile ../../ --load
	BUILD_COMMAND_ARM64=docker buildx build --platform linux/arm64 -t $(IMAGE_NAME) -f Dockerfile ../../ --load
	BUILD_COMMAND_ALL=docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE_NAME) -f Dockerfile ../../ --push
	PUSH_COMMAND=docker push $(IMAGE_NAME)

.PHONY: dev prod push all
dev:
	$(BUILD_COMMAND_ARM64)
	$(PUSH_COMMAND)

prod:
	$(BUILD_COMMAND)
	$(PUSH_COMMAND)

push:
	$(PUSH_COMMAND)

all:
	$(BUILD_COMMAND_ALL)

