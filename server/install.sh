#!/bin/sh
echo "Remove old binaries..."
rm main
rm rbdns-server

echo "Creating new binaries..."
go build main.go
mv main rbdns-server

echo "moving rbdns-server to $GOPATH/bin..."
mv rbdns-server $GOPATH/bin 