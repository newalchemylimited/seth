#!/bin/sh
cd ..
go generate . && go install . 
cd test
go generate . && go run main.go generated.go