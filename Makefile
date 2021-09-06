REPOSITORY ?= ghcr.io/jlandowner/envoy-ext-authz-line
TAG ?= local-build

all: build

build:
	docker build . -t $(REPOSITORY):$(TAG)
	docker tag $(REPOSITORY):$(TAG) $(REPOSITORY):latest