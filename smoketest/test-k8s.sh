#!/bin/bash

VERSION=$1

if [ -z "${VERSION}" ]; then
	echo "usage: $(basename $0) <version>"
	echo "version = 1.x.x"
	exit 1;
fi;

./get-kubectl.sh ${VERSION}

rm -f bin/kubectl

chmod +x bin/kubectl-${VERSION}

cd bin

ln -s kubectl-${VERSION} kubectl

cd - 


minikube delete

minikube start --kubernetes-version v${VERSION}

# remove taint on master so that we can schedule pods on it
bin/kubectl taint nodes minikube node-role.kubernetes.io/master:NoSchedule-

