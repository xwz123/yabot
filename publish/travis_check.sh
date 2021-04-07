#!/bin/bash

LATEST_STABLE_SUPPORTED_GO_VERSION="1.13"

main() {
  if local_go_version_is_latest_stable
  then
    run_gofmt
    run_golint
  fi
  run_build
}

local_go_version_is_latest_stable() {
  go version | grep  $LATEST_STABLE_SUPPORTED_GO_VERSION
}

log_error() {
   echo -e "\x1B[91m${@}\x1B[0m" 1>&2
}

run_gofmt() {
  GOFMT_FILES=$(gofmt -l .)
  if [ -n "$GOFMT_FILES" ]
  then
    log_error "gofmt failed for the following files:
$GOFMT_FILES
please run 'gofmt -w .' on your changes before committing."
#    exit 1
  fi
}

run_golint() {
  GOLINT_ERRORS=$(golint ./... | grep -v "Id should be")
  if [ -n "$GOLINT_ERRORS" ]
  then
    log_error "golint failed for the following reasons:
$GOLINT_ERRORS
please run 'golint ./...' on your changes before committing."
#    exit 1
  fi
}


run_unit_tests() {
  if [ -z "$NOTEST" ]
  then
    log_error 'Running short tests...'
    env AMQP_URL= go test -short
  fi
}

run_build(){
    bazel build //gitee/cmd/hook:hook
}

main
