#!/bin/sh
echo "Remove old binaries..."
rm main
rm rbdns-client

echo "Creating new binaries..."
go build main.go
mv main rbdns-client

echo "moving rbdns-client to $GOPATH/bin..."
mv rbdns-client $GOPATH/bin 