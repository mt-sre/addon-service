SHELL=/bin/bash
.SHELLFLAGS=-euo pipefail -c

UNAME_OS:=$(shell uname -s)
UNAME_OS_LOWER:=$(shell uname -s | awk '{ print tolower($$0); }') # UNAME_OS but in lower case
UNAME_ARCH:=$(shell uname -m)

# Dependency version
BUF_VERSION:=1.3.1

# PATH/Bin
PROJECT_DIR:=$(shell pwd)
DEPENDENCIES:=.deps
DEPENDENCY_BIN:=$(abspath $(DEPENDENCIES)/bin)
DEPENDENCY_VERSIONS:=$(abspath $(DEPENDENCIES)/$(UNAME_OS)/$(UNAME_ARCH)/versions)
export PATH:=$(DEPENDENCY_BIN):$(PATH)


# TODO: move this to devkube and use magefile!
proto-gen:
	mkdir -p ${DEPENDENCY_BIN};
	curl -sSL \
		"https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-${UNAME_OS}-${UNAME_ARCH}" \
		-o "${DEPENDENCY_BIN}/buf" && \
	chmod +x "${DEPENDENCY_BIN}/buf";
	${DEPENDENCY_BIN}/buf generate;
	@go mod tidy;