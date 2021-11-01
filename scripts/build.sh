#!/bin/bash

docker build . -t krinder:latest

echo $GITHUB_PAT | docker login ghcr.io -u ddouglas --password-stdin

IMAGE_BASE="ghcr.io/eveisesi/krinder"
TAG_NAME=$(git describe --tags $(git rev-list --tags --max-count=1))
VERSION=$(echo $TAG_NAME | sed -e 's,.*/\(.*\),\1,' | sed -e 's/^v//')
IMAGE_NAME=$IMAGE_BASE
echo "Image Base: $IMAGE_BASE"
echo "Tagging and Pushing $IMAGE_NAME:$VERSION"
docker tag krinder:latest $IMAGE_NAME:$VERSION
docker push $IMAGE_NAME:$VERSION
echo "Tagging and Push $IMAGE_NAME:latest"
docker tag krinder:latest $IMAGE_NAME:latest
docker push $IMAGE_NAME:latest

clear