#!/bin/bash
GOOS=js GOARCH=wasm go build -o static/main.wasm main.go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./static