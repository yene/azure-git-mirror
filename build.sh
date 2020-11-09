#!/bin/bash
set -e
env GOOS=linux GOARCH=amd64 go build -o git-mirror-linux
upx git-mirror-linux
env GOOS=darwin GOARCH=amd64 go build -o git-mirror-macOS
upx git-mirror-macOS
