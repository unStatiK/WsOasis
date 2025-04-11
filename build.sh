#!/bin/bash

gofmt -w oasis.go
go build -mod=vendor -ldflags="-w -s" -a -gcflags=all="-l -B" oasis.go