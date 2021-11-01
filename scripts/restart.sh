#!/bin/bash

docker compose up -d mongo redis
sleep 10
docker compose up -d app