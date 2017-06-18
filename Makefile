REPO=golang
TAG=1.8-alpine3.6
BIN=k8s-app-config
VERSION?=dev
GOOS=linux
GOARCH=amd64

.PHONY: build pull fmt shell

build: pull
	docker run --rm -v $(PWD):/go/src/$(BIN) -w /go/src/$(BIN) $(REPO):$(TAG) \
	sh -c 'GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags "-s -w -X main.VERSION=$(VERSION)" -v'

pull:
	docker images | grep '$(REPO)' | grep '$(TAG)' > /dev/null; \
	if [ $$? -ne 0 ]; then \
	docker pull $(REPO):$(TAG); \
	fi

fmt:
	@docker run --rm -v $(PWD):/go/src/$(BIN) -w /go/src/$(BIN) $(REPO):$(TAG) go fmt

shell: pull
	@docker run --rm -ti --init -v $(PWD):/go/src/$(BIN) -w /go/src/$(BIN) $(REPO):$(TAG) /bin/sh
