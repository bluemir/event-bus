VERSION?=$(shell git describe --long --tags --dirty --always)
export VERSION

IMPORT_PATH=$(shell cat go.mod | head -n 1 | awk '{print $$2}')
BIN_NAME=$(notdir $(IMPORT_PATH))

export GO111MODULE=on
export GIT_TERMINAL_PROMPT=1

DOCKER_IMAGE_NAME=bluemir/$(BIN_NAME)

## Go Sources
GO_SOURCES = $(shell find . -name "vendor"  -prune -o \
                            -type f -name "*.go" -print)

## FE sources
JS_SOURCES    = $(shell find app/js       -type f -name '*.js'   -print)
HTML_SOURCES  = $(shell find app/html     -type f -name '*.html' -print)
CSS_SOURCES   = $(shell find app/css      -type f -name '*.css'  -print)
WEB_LIBS      = $(shell find app/lib      -type f                -print)


DISTS =
DISTS += $(JS_SOURCES:app/js/%=build/dist/js/%)
DISTS += $(CSS_SOURCES:app/css/%=build/dist/css/%)
DISTS += $(WEB_LIBS:app/lib/%=build/dist/lib/%)

default: build

## Web dist
build/dist/%: app/%
	@mkdir -p $(dir $@)
	cp $< $@

build: build/$(BIN_NAME)

build/$(BIN_NAME).unpacked: $(GO_SOURCES) makefile
	@mkdir -p build
	go build -v \
		-trimpath \
		-ldflags "\
			-X main.AppName=$(BIN_NAME) \
			-X main.Version=$(VERSION)  \
		" \
		$(OPTIONAL_BUILD_ARGS) \
		-o $@ main.go
build/$(BIN_NAME): build/$(BIN_NAME).unpacked $(HTML_SOURCES) $(DISTS)
	@mkdir -p build
	cp $< $@.tmp
	rice append -v \
		-i $(IMPORT_PATH)/pkg/dist \
		--exec $@.tmp
	mv build/$(BIN_NAME).tmp $@

docker: build/docker-image

build/docker-image: build/Dockerfile $(GO_SOURCES) $(HTML_SOURCES) $(DISTS)
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE_NAME):$(VERSION) \
		-f $< .
	echo $(DOCKER_IMAGE_NAME):$(VERSION) > $@

build/Dockerfile: export BIN_NAME:=$(BIN_NAME)
build/Dockerfile: Dockerfile.template
	@mkdir -p build
	cat $< | envsubst '$${BIN_NAME}' > $@

push: build/docker-image.pushed

build/docker-image.pushed: build/docker-image
	docker push $(shell cat build/docker-image)
	echo $(shell cat build/docker-image) > $@

clean:
	rm -rf build/
	ps -f -C make | grep "test run" | awk '{print $$2}' | xargs kill || true
	ps -f -C $(BIN_NAME) | grep 'retry=10' | awk '{print $$2}' | xargs kill || true

run: build/$(BIN_NAME)
	$< -vv --bind=:3003 --network=test --key="" --peer "ws://localhost:3003/v1/stream"

auto-run:
	while true; do \
		$(MAKE) .sources | \
		entr -rd $(MAKE) test run ;  \
		echo "hit ^C again to quit" && sleep 1  \
	; done

.sources:
	@echo \
	makefile \
	$(GO_SOURCES) \
	$(JS_SOURCES) \
	$(HTML_SOURCES) \
	$(CSS_SOURCES) \
	$(WEB_LIBS) \
	| tr " " "\n"

test:
	go test -v ./pkg/...

helper: build/$(BIN_NAME)
	ps -f -C $(BIN_NAME) | grep 'retry=10' | awk '{print $$2}' | xargs kill || true;
	for port in $$(seq 8021 8025) ; do \
		$< -vv \
			--retry=10 \
			--bind=:$$port \
			--network=test \
			--key="" \
			--peer=ws://localhost:3003/v1/stream & \
	done

.PHONY: build docker push clean run auto-run .sources test deploy
