#!/usr/bin/env bash

# update.sh updates tests golden files

set -euo pipefail

go test ./examples ./pkg/gmg --update
