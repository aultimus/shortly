#!/bin/bash

sudo docker run -p 8080:8080/tcp --env-file .env.local 361313012007.dkr.ecr.us-west-2.amazonaws.com/shortly:latest

#sudo docker compose up
# for debugging
# sudo docker run -it --entrypoint /bin/sh aultimus/shortly:latest
# sudo docker exec -i -t $CONTAINER_ID /bin/bash
