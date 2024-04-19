#!/bin/bash

echo $1 # commit hash
echo $2 # version

CGO_ENABLED=1 go build -ldflags="-s -w -X 'imgu2/services.GIT_COMMIT=$1' -X 'imgu2/services.VERSION=$2'" -trimpath -o imgu2
