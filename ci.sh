#!/usr/bin/env bash

golangci-lint run --fix
go fmt ./...
