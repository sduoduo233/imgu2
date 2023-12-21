#!/bin/bash

# testing: docker run -p 3000:3000 --rm -v ./:/data -it alpine:latest /bin/sh


GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w -extldflags=-static" -trimpath -o build/imgu2-linux-amd64
