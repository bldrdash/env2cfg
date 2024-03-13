BINARY=env2cfg
CHECKS=all

ifeq ($(VERSION),)
	VERSION = $(shell git tag | sort -V | tail -1)
endif

RELEASE=$(shell git rev-parse HEAD)
DATE=$(shell date "+%Y-%m-%d_%I:%M:%S%p")
RELEASEFLAGS=-s -w -X main.BuildTime=${DATE} -X main.Githash=${RELEASE}

.PHONY: build run test cover status tidy lint version watch check-git-clean bump release example

build:
	@go build -trimpath -o "${BINARY}"

run: build
	@${BINARY}

test:
	go test

cover:
	go test -cover -v

status:
	@git status

tidy:
	@go mod tidy

lint:
	@staticcheck --checks ${CHECKS}

version:
	@echo $(VERSION)

watch:
	@air

config.yaml: .env
	@./$(BINARY) .env config.tpl.yaml $@

config: config.yaml

check-git-clean:
	@git diff --quiet	


bump: #check-git-clean
	$(eval NEWVER=$(shell grep -oP '^var\s+version\s*=\s*"v\K[^"]+?(?=")' version.go |awk -F. '{OFS="."; $$NF+=1; print $0}'))
	@echo $(NEWVER)
	sed -i -e '/^var version /s/"[^"][^"]*"/\"v$(NEWVER)\"/' version.go

release: tidy lint
	@go build -ldflags="${RELEASEFLAGS} -X main.Version=${VERSION}" -trimpath -o ${BINARY}

example: build
	./env2cfg -I example/config.tmpl.yaml example/dotenv config.yaml
	@cat config.yaml
	@rm config.yaml	