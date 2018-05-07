#!/bin/sh

export GOPATH=/go/src/.go
go get -v
go run /go/src/main.go
