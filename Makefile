#!make
OS ?= darwin
ARCH ?= amd64
PREFIX ?= ./build
VERSION := 1.11
TAG = $(shell date -u +'%Y.%m.%d-%H')
ORG = callowaylc
REPO = mq
ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

-include .env
export

.PHONY: make install release test clean

make:
	mkdir -p ./build
	#rm -rf ./build/*

	- test "$(OS)" = "linux" && docker run \
			--rm \
			-e GOOS=$(OS) \
			-e GOARCH=$(ARCH) \
			-v `pwd`:/opt/bin \
			-v `pwd`/build/cache:/go/pkg \
			-w /opt/bin golang:$(VERSION) \
				go build -v -o ./release/mq-$(OS)-$(ARCH) ./cmd/mq.go

	- test "$(OS)" = "darwin" && \
			vgo build -v -o ./release/mq-$(OS)-$(ARCH) ./cmd/mq.go


install:
	mv ./build/logger-$(OS)-$(ARCH) $(PREFIX)/logger

release:
	# NOTE: Add latest along with calendar version
	# NOTE: Remove other releases from the same day
	mkdir -p ./release
	rm -rf ./release/*

	OS=darwin make & \
	OS=linux make & \
	wait

	git tag $(TAG) -f
	git push origin $(TAG) -f

	- github-release delete \
		--user $(ORG) \
		--repo $(REPO) \
		--tag $(TAG)

	github-release release --draft \
		--user $(ORG) \
		--repo $(REPO) \
		--tag $(TAG) \
		--name $(TAG)

	github-release release --draft \
		--user $(ORG) \
		--repo $(REPO) \
		--tag latest \
		--name $(TAG)

	ls ./release/* | xargs -n1 basename | xargs -n1 -I{} github-release upload \
		--replace \
		--user $(ORG) \
		--repo $(REPO) \
		--tag $(TAG) \
		--name {} \
    --file ./release/{}

	ls ./release/* | xargs -n1 basename | xargs -n1 -I{} github-release upload \
		--replace \
		--user $(ORG) \
		--repo $(REPO) \
		--tag latest \
		--name {} \
    --file ./release/{}

publish:
	github-release edit \
		--user $(ORG) \
		--repo $(REPO) \
		--tag $(TAG) \
		--name $(TAG)
	github-release edit \
		--user $(ORG) \
		--repo $(REPO) \
		--tag $(TAG) \
		--name latest

test:
	vgo build -v -o ./build/mq ./cmd/mq.go

clean:
	rm -rf ./build

%:
	@:
