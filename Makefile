SHELL := /bin/bash
PACKAGE := "github.com/thrisp/deaffink"
TARGET := $(shell echo $${PWD\#\#*/})
.DEFAULT_GOAL: $(TARGET)
# These will be provided to the target
VERSIONTAG := "0.0.1"
VERSIONHASH := `git rev-parse HEAD`
VERSIONDATE := `date -u +%d-%m-%Y.%H:%M:%S`

LDFLAGS=-ldflags "-X=main.versionTag=$(VERSIONTAG) -X=main.versionHash=$(VERSIONHASH) -X=main.versionDate=$(VERSIONDATE)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build clean install uninstall fmt simplify check run

all: check install

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET)
	@true

clean:
	@rm -f $(TARGET)

install:
	@go install $(LDFLAGS)

install-binary:
	@cp $(TARGET) $(GOPATH)/bin/$(TARGET) 

uninstall: clean
	@rm -f $$(which ${TARGET})

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

check:
	@test -z $(shell gofmt -l main.go | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@go tool vet ${SRC}

run: install
	@$(TARGET)
