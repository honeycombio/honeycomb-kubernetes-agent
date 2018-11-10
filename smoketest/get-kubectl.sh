#!/bin/bash

VERSION=$1

if [ -z "${VERSION}" ]; then
	echo "usage: $(basename $0) <version>"
	echo "version = 1.x.x"
	exit 1;
fi;

if [ "$(uname)" == "Darwin" ]; then
	curl -L https://storage.googleapis.com/kubernetes-release/release/v${VERSION}/bin/darwin/amd64/kubectl -o bin/kubectl-${VERSION}
else
	curl -L https://storage.googleapis.com/kubernetes-release/release/v${VERSION}/bin/linux/amd64/kubectl -o bin/kubectl-${VERSION}

fi;
