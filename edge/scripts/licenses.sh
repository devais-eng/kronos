#!/usr/bin/env bash

# This script generates a CSV file containing all dependencies
# and bundles licenses that needs to be redistributed.

# This script requires the go-licenses tool
# https://github.com/google/go-licenses

set -e


go-licenses csv ./... 2> /dev/null | grep -v "devais.it" > ./licenses.csv
go-licenses save --save_path=./licenses ./... 2> /dev/null
