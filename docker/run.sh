#!/bin/bash


sudo systemctl start docker

docker run -p 6390:6379 -v $PWD/data:/data -d redis redis-server --appendonly yes
echo "Redis is listening on port 6390. Data is stored inside 'data' directory"
