#!/bin/bash

gofmt -w oasis.go
go build -ldflags="-w -s" -a -gcflags=all="-l -B" oasis.go
