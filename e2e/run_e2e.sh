#!/bin/bash

set -o nounset
set -o pipefail

echo "SETUP install managed-rbac"
kubectl config use-context kind-hub
kubectl apply -f config/crd/bases/rbac.open-cluster-management.io_managedrbacs.yaml
kubectl apply -f config/rbac
kubectl apply -f config/deploy
if kubectl wait --for=condition=available --timeout=600s deployment/managed-rbac -n open-cluster-management; then
    echo "Deployment available"
else
    echo "Deployment not available"
    exit 1
fi

echo "TEST ManagedRBAC"
kubectl config use-context kind-hub
kubectl apply -f config/samples/rbac.open-cluster-management.io_v1alpha1_managedrbac.yaml -n cluster1
sleep 10
work_kubectl_command=$(kubectl -n cluster1 get managedrbac -o yaml | grep kubectl | grep ManifestWork)
if $work_kubectl_command; then
    echo "ManifestWork found"
else
    echo "ManifestWork not found"
    exit 1
fi

kubectl config use-context kind-cluster1
if kubectl -n default get role managedrbac-sample; then
    echo "managedrbac-sample role found"
else
    echo "managedrbac-sample role not found"
    exit 1
fi
if kubectl -n default get rolebinding managedrbac-sample; then
    echo "managedrbac-sample rolebinding found"
else
    echo "managedrbac-sample rolebinding not found"
    exit 1
fi
if kubectl get clusterrole managedrbac-sample; then
    echo "managedrbac-sample clusterrole found"
else
    echo "managedrbac-sample clusterrole not found"
    exit 1
fi
if kubectl get clusterrolebinding managedrbac-sample; then
    echo "managedrbac-sample clusterrolebinding found"
else
    echo "managedrbac-sample clusterrolebinding not found"
    exit 1
fi
