#!/bin/sh
go generate github.com/newalchemylimited/seth/bindgen && go install github.com/newalchemylimited/seth/bindgen && go generate . && go run main.go generated.go