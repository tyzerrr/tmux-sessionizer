#!/usr/bin/env bash

golangci-lint run --fix
go test ./... -v
go fmt ./...
