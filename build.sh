#!/bin/bash

rm settler
go clean ./go-src/
go build -o ./settler ./go-src/
