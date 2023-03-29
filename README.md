# ManagedRBAC
`ManagedRBAC` is an OCM custom resource for helping the administrators to automatically distribute RBAC resources to the managed clusters and manage the lifecycle of those RBAC resources including Role, ClusterRole, RoleBinding, ClusterRoleBinding.
This project is intended to enhance the usability of [ManagedServiceAccount](https://github.com/open-cluster-management-io/managed-serviceaccount). The ManagedServiceAccount can help users with authentication fleet management and the ManagedRBAC can cover the authorization aspects of fleet management.

## Description
This repo contains the API definition and controller of `ManagedRBAC`.
A valid ManagedRBAC resource should be in a "cluster namespace" and the associated RBAC resources will be deliver to the associated managed cluster with that "cluster namespace". The ManagedRBAC controller using ManifestWork API will ensuring, updating, removing the RBAC resources from each of the managed cluster.The ManagedRBAC API will protect these distributed RBAC resources from unexpected modification and removal. In additional to the typical allowable RBAC binding subject resources (Group, ServiceAccount, and User), `ManagedServiceAccount` can be used as a subject as well. If the subject of the binding is a ManagedServiceAccount then the controller will calculate and create RBAC resources with the ServiceAccount that is handled by the ManagedServiceAccount.

## Dependencies
- The Open Cluster Management (OCM) multi-cluster environment needs to be setup. See [OCM website](https://open-cluster-management.io/) on how to setup the environment.
- Optional: [ManagedServiceAccount](https://github.com/open-cluster-management-io/managed-serviceaccount) add-on installed if you want to leverage `ManagedServiceAccount` resource as RBAC subject.

## Getting Started
1. Setup an OCM Hub cluster and registered an OCM Managed cluster. See [Open Cluster Management Quick Start](https://open-cluster-management.io/getting-started/quick-start/) for more details.

2. On the Hub cluster, install the `ManagedRBAC` API and run the controller,
```
cd managed-rbac/
make install
make run
```

3. On the Hub cluster, apply the sample (modify the cluster1 namespace to your managed cluster name):
```
kubectl -n cluster1 apply -f config/samples/rbac.open-cluster-management.io_v1alpha1_managedrbac.yaml
kubectl -n cluster1 get managedrbac -o yaml
...
  status:
    conditions:
    - lastTransitionTime: "2023-04-12T15:19:04Z"
      message: |-
        Run the following command to check the ManifestWork status:
        kubectl -n cluster1 get ManifestWork managedrbac-sample-f15f0 -o yaml
      reason: AppliedRBACManifestWork
      status: "True"
      type: AppliedRBACManifestWork
```

4. On the Managed cluster, check the RBAC resources
```
kubectl -n default get role
NAME                 CREATED AT
managedrbac-sample   2023-04-12T15:19:04Z
```

## Community, discussion, contribution, and support

Check the [CONTRIBUTING Doc](CONTRIBUTING.md) for how to contribute to the repo.

### Communication channels

Slack channel: [#open-cluster-mgmt](https://kubernetes.slack.com/channels/open-cluster-mgmt)

## License

This code is released under the Apache 2.0 license. See the file [LICENSE](LICENSE) for more information.
