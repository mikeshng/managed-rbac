/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	workv1 "open-cluster-management.io/api/work/v1"
	mrbacv1alpha1 "open-cluster-management.io/managed-rbac/api/v1alpha1"
)

// generateManifestWorkName returns the ManifestWork name for a given ManagedRBAC.
// It uses the ManagedRBAC name with the suffix of the first 5 characters of the UID
func generateManifestWorkName(managedRBAC mrbacv1alpha1.ManagedRBAC) string {
	return managedRBAC.Name + "-" + string(managedRBAC.UID)[0:5]
}

// buildManifestWork wraps the payloads in a ManifestWork
func buildManifestWork(managedRBAC mrbacv1alpha1.ManagedRBAC, manifestWorkName string,
	clusterRole *rbacv1.ClusterRole,
	clusterRoleBinding *rbacv1.ClusterRoleBinding,
	roles []rbacv1.Role,
	roleBindings []rbacv1.RoleBinding) *workv1.ManifestWork {
	var manifests []workv1.Manifest

	if clusterRole != nil {
		manifests = append(manifests, workv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRole}})
	}

	if clusterRoleBinding != nil {
		manifests = append(manifests, workv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRoleBinding}})
	}

	if len(roles) > 0 {
		for i := range roles {
			manifests = append(manifests, workv1.Manifest{RawExtension: runtime.RawExtension{Object: &roles[i]}})
		}
	}

	if len(roleBindings) > 0 {
		for i := range roleBindings {
			manifests = append(manifests, workv1.Manifest{RawExtension: runtime.RawExtension{Object: &roleBindings[i]}})
		}
	}

	// setup the owner so when ManagedRBAC is deleted, the associated ManifestWork is also deleted
	owner := metav1.NewControllerRef(&managedRBAC, mrbacv1alpha1.GroupVersion.WithKind("ManagedRBAC"))

	return &workv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            manifestWorkName,
			Namespace:       managedRBAC.Namespace,
			OwnerReferences: []metav1.OwnerReference{*owner},
		},
		Spec: workv1.ManifestWorkSpec{
			Workload: workv1.ManifestsTemplate{
				Manifests: manifests,
			},
		},
	}
}
