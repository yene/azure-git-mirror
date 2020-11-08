#!/bin/bash
set -e
go build -o git-mirror
upx git-mirror
