#!/usr/bin/env sh
set -eu
docker build -f benzhi.Dockerfile -t biotreat:local .
