#!/bin/bash

wget https://dl.google.com/go/go1.10.linux-amd64.tar.gz

sudo tar -xvf go1.10.linux-amd64.tar.gz

sudo mv go /usr/local

mkdir ~/go-packages/

export GOROOT="/usr/local/go"

export GOPATH="$HOME/go-packages"

export PATH="$GOPATH/bin:$GOROOT/bin:$PATH"
