#!/bin/bash

CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -o imgu2
