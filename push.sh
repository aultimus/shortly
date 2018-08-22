#!/bin/bash
sudo $(aws ecr get-login --no-include-email --region us-west-2)
sudo docker push 361313012007.dkr.ecr.us-west-2.amazonaws.com/shortly
