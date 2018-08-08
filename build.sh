#!/bin/bash
pushd shortly
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .
chmod +x shortly
popd
sudo docker build -t aultimus/shortly:latest .
sudo docker run aultimus/shortly:latest

# for debugging
# sudo docker run -it --entrypoint /bin/sh aultimus/shortly:latest
