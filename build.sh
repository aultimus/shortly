#!/bin/bash
pushd shortly
SHA=`git rev-parse --short HEAD`
CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -X main.gitSHA=$SHA -extldflags '-static'" .
chmod +x shortly
popd
sudo docker build -t 361313012007.dkr.ecr.us-west-2.amazonaws.com/shortly:latest .
