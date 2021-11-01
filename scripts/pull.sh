#!/bin/bash

if [ -z "$GITHUB_PAT" ]; then echo "GITHUB_PAT env must be set to run this script"; exit 1; fi;

echo $GITHUB_PAT | docker login ghcr.io -u ddouglas --password-stdin

docker compose pull

clear