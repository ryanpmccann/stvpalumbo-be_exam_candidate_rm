## watch for trailing spaces!
THIS_MAKEFILE = $(lastword $(MAKEFILE_LIST))
REPOROOT = $(abspath  $(dir $(THIS_MAKEFILE)))

export GOPATH := $(REPOROOT)

ifdef GOROOT
	#$(error Define GOROOT prior to running make)
	PATH := $(GOROOT)/bin:$(PATH)
endif

GOCMD ?= go
GOCLEAN := $(GOCMD) clean

.PHONY: all
all: fileconverter

.PHONY: clean
clean:
	@rm -rf $(REPOROOT)/bin
	@rm -rf $(REPOROOT)/pkg
	@rm -rf $(REPOROOT)/src/github.com
	@rm -rf $(REPOROOT)/src/golang.org

.PHONY: test
test: fileconverter_test 

fileconverter_test fileconverter: $(REPOROOT)/bin/get_deps report_go_version

$(REPOROOT)/bin/get_deps:
	$(GOCMD) get github.com/fsnotify/fsnotify
	$(GOCMD) get github.com/golang/glog
	mkdir -p $(REPOROOT)/bin
	touch $(REPOROOT)/bin/get_deps

report_go_version:
	@echo "current go version:"; $(GOCMD) version

fileconverter_test:
	cd $(REPOROOT)/src/fileconverter; $(GOCMD) test

fileconverter:
	cd $(REPOROOT)/src/fileconverter; $(GOCMD) install
