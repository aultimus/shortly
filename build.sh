#!/bin/bash
pushd shortly
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .
chmod +x shortly
popd
sudo docker build -t 361313012007.dkr.ecr.us-west-2.amazonaws.com/shortly:latest .
