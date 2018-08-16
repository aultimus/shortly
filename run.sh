#!/bin/bash

sudo docker run -p 8080:8080/tcp -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY aultimus/shortly:latest

#sudo docker compose up
# for debugging
# sudo docker run -it --entrypoint /bin/sh aultimus/shortly:latest
# sudo docker exec -i -t $CONTAINER_ID /bin/bash