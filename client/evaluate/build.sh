#!/bin/sh
echo "Remove old evaluate binaries..."
rm main
rm evaluate

echo "Build new evaluate binaries..."
go build main.go
mv main evaluate