#!/bin/bash

go build -ldflags="-w -s" -a -gcflags=all="-l -B" oasis.go
