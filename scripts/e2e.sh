#!/usr/bin/env bash

export PATH=$GOPATH/bin:./bin:$PATH

mkdir -p ~/.kube
oi-local-cluster cluster create | oi cluster inject --name test | oi node inject --name test --cluster test --role controlplane | oi node inject --name loadbalancer --cluster test --role controlplane-ingress | tee cluster.txt | oi reconcile
cat cluster.txt | oi cluster kubeconfig --cluster test --endpoint-host-override 127.0.0.1 > ~/.kube/config
docker ps -a

RETRIES=0
MAX_RETRIES=5
until kubectl cluster-info &> /dev/null || [ $RETRIES -eq $MAX_RETRIES ]; do
   sleep 1
done

kubectl cluster-info
kubectl version
