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
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	workv1 "open-cluster-management.io/api/work/v1"
	mrbacv1alpha1 "open-cluster-management.io/managed-rbac/api/v1alpha1"
	msav1alpha1 "open-cluster-management.io/managed-serviceaccount/api/v1alpha1"
	msacommon "open-cluster-management.io/managed-serviceaccount/pkg/common"
)

var _ = Describe("ManagedRBAC controller", func() {

	const (
		clusterName = "cluster1"
		mbacName    = "managedrbac-1"
		msaName     = "managedsa-1"
	)

	mbacKey := types.NamespacedName{Name: mbacName, Namespace: clusterName}
	ctx := context.Background()

	Context("When ManagedRBAC is created/updated/deleted", func() {
		It("Should create/update/delete ManifestWork", func() {
			By("Creating the OCM ManagedCluster")
			managedCluster := clusterv1.ManagedCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterName,
				},
			}
			Expect(k8sClient.Create(ctx, &managedCluster)).Should(Succeed())

			By("Creating the OCM ManagedCluster namespace")
			managedClusterNs := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   clusterName,
					Labels: map[string]string{"a": "b"},
				},
			}
			Expect(k8sClient.Create(ctx, &managedClusterNs)).Should(Succeed())

			By("Creating the ManagedRBAC")
			managedRBAC := mrbacv1alpha1.ManagedRBAC{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mbacName,
					Namespace: clusterName,
				},
				Spec: mrbacv1alpha1.ManagedRBACSpec{
					ClusterRole: &mrbacv1alpha1.ClusterRole{
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{"apps"},
							Resources: []string{"services"},
							Verbs:     []string{"create"},
						}},
					},
					Roles: &[]mrbacv1alpha1.Role{{
						Namespace: "default",
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{""},
							Resources: []string{"namespaces"},
							Verbs:     []string{"get"},
						}},
					}},
					ClusterRoleBinding: &mrbacv1alpha1.ClusterRoleBinding{
						Subject: rbacv1.Subject{
							Kind:      "ServiceAccount",
							Name:      "klusterlet-work-sa",
							Namespace: "open-cluster-management-agent",
						},
					},
					RoleBindings: &[]mrbacv1alpha1.RoleBinding{
						{
							Namespace: "default",
							RoleRef:   mrbacv1alpha1.RoleRef{Kind: "ClusterRole"},
							Subject: rbacv1.Subject{
								Kind:      "ServiceAccount",
								Name:      "klusterlet-work-sa",
								Namespace: "open-cluster-management-agent",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, &managedRBAC)).Should(Succeed())
			mwKey := types.NamespacedName{Name: generateManifestWorkName(managedRBAC), Namespace: clusterName}
			mw := workv1.ManifestWork{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())

			By("Updating the ManagedRBAC")
			time.Sleep(3 * time.Second)
			oldRv := mw.GetResourceVersion()
			Expect(k8sClient.Get(ctx, mbacKey, &managedRBAC)).Should(Succeed())
			managedRBAC.Spec.Roles = nil
			Eventually(func() bool {
				if err := k8sClient.Update(ctx, &managedRBAC); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return false
				}
				return oldRv != mw.GetResourceVersion()
			}).Should(BeTrue())

			By("Deleting the ManagedRBAC")
			Expect(k8sClient.Get(ctx, mbacKey, &managedRBAC)).Should(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Delete(ctx, &managedRBAC); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, mbacKey, &managedRBAC); err != nil {
					return true
				}
				return false
			}).Should(BeTrue())

			By("Creating the ManagedServiceAccount")
			managedSA := msav1alpha1.ManagedServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: clusterName,
					Name:      msaName,
				},
			}
			Expect(k8sClient.Create(ctx, &managedSA)).Should(Succeed())

			By("Creating the ManagedServiceAccount addon")
			saAddon := addonv1alpha1.ManagedClusterAddOn{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: clusterName,
					Name:      msacommon.AddonName,
				},
			}
			Expect(k8sClient.Create(ctx, &saAddon)).Should(Succeed())

			By("Creating the ManagedRBAC with subject ManagedServiceAccount")
			managedRBAC = mrbacv1alpha1.ManagedRBAC{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mbacName,
					Namespace: clusterName,
				},
				Spec: mrbacv1alpha1.ManagedRBACSpec{
					ClusterRole: &mrbacv1alpha1.ClusterRole{
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{"apps"},
							Resources: []string{"services"},
							Verbs:     []string{"create"},
						}},
					},
					Roles: &[]mrbacv1alpha1.Role{{
						Namespace: "default",
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{""},
							Resources: []string{"namespaces"},
							Verbs:     []string{"get"},
						}},
					}},
					ClusterRoleBinding: &mrbacv1alpha1.ClusterRoleBinding{
						Subject: rbacv1.Subject{
							APIGroup: "authentication.open-cluster-management.io",
							Kind:     "ManagedServiceAccount",
							Name:     msaName,
						},
					},
					RoleBindings: &[]mrbacv1alpha1.RoleBinding{
						{
							Namespace: "default",
							RoleRef:   mrbacv1alpha1.RoleRef{Kind: "ClusterRole"},
							Subject: rbacv1.Subject{
								APIGroup: "authentication.open-cluster-management.io",
								Kind:     "ManagedServiceAccount",
								Name:     msaName,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, &managedRBAC)).Should(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())

			By("Deleting the ManagedRBAC")
			Expect(k8sClient.Delete(ctx, &managedRBAC)).Should(Succeed())

			By("Creating the ManagedRBAC with namespaceSelector")
			managedRBAC = mrbacv1alpha1.ManagedRBAC{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mbacName,
					Namespace: clusterName,
				},
				Spec: mrbacv1alpha1.ManagedRBACSpec{
					ClusterRole: &mrbacv1alpha1.ClusterRole{
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{"apps"},
							Resources: []string{"services"},
							Verbs:     []string{"create"},
						}},
					},
					Roles: &[]mrbacv1alpha1.Role{{
						NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "b"}},
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{""},
							Resources: []string{"namespaces"},
							Verbs:     []string{"get"},
						}},
					}},
					RoleBindings: &[]mrbacv1alpha1.RoleBinding{
						{
							NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "b"}},
							RoleRef:           mrbacv1alpha1.RoleRef{Kind: "ClusterRole"},
							Subject: rbacv1.Subject{
								APIGroup: "authentication.open-cluster-management.io",
								Kind:     "ManagedServiceAccount",
								Name:     msaName,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, &managedRBAC)).Should(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())

			k8sClient.Get(ctx, mbacKey, &managedRBAC)
			managedRBAC.Spec = mrbacv1alpha1.ManagedRBACSpec{
				ClusterRole: &mrbacv1alpha1.ClusterRole{
					Rules: []rbacv1.PolicyRule{{
						APIGroups: []string{"apps"},
						Resources: []string{"deployments"},
						Verbs:     []string{"create"},
					}},
				}}
			Expect(k8sClient.Update(ctx, &managedRBAC)).Should(Succeed())

			By("Deleting the ManagedRBAC")
			Expect(k8sClient.Delete(ctx, &managedRBAC)).Should(Succeed())
		})
	})

	Context("When ManagedRBAC is created without prereqs", func() {
		It("Should have error status", func() {
			By("Creating the ManagedRBAC with no rules")
			managedRBAC := mrbacv1alpha1.ManagedRBAC{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mbacName,
					Namespace: clusterName,
				},
			}

			Expect(k8sClient.Create(ctx, &managedRBAC)).Should(Succeed())
			mwKey := types.NamespacedName{Name: generateManifestWorkName(managedRBAC), Namespace: clusterName}
			mw := workv1.ManifestWork{}
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return true
				}
				return false
			}).Should(BeTrue())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mbacKey, &managedRBAC); err != nil {
					return false
				}
				if len(managedRBAC.Status.Conditions) > 0 {
					return managedRBAC.Status.Conditions[0].Type == mrbacv1alpha1.ConditionTypeValidation &&
						managedRBAC.Status.Conditions[0].Status == metav1.ConditionFalse
				}
				return false
			}).Should(BeTrue())

			By("Deleting the ManagedRBAC")
			Expect(k8sClient.Delete(ctx, &managedRBAC)).Should(Succeed())

			By("Creating the ManagedRBAC with invalid ClusterRoleBinding ManagedServiceAccount")
			managedRBAC = mrbacv1alpha1.ManagedRBAC{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mbacName,
					Namespace: clusterName,
				},
				Spec: mrbacv1alpha1.ManagedRBACSpec{
					ClusterRole: &mrbacv1alpha1.ClusterRole{
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{"apps"},
							Resources: []string{"services"},
							Verbs:     []string{"create"},
						}},
					},
					ClusterRoleBinding: &mrbacv1alpha1.ClusterRoleBinding{
						Subject: rbacv1.Subject{
							APIGroup: "authentication.open-cluster-management.io",
							Kind:     "ManagedServiceAccount",
							Name:     msaName + "1",
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, &managedRBAC)).Should(Succeed())
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return true
				}
				return false
			}).Should(BeTrue())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mbacKey, &managedRBAC); err != nil {
					return false
				}
				if len(managedRBAC.Status.Conditions) > 0 {
					return managedRBAC.Status.Conditions[0].Type == mrbacv1alpha1.ConditionTypeAppliedRBACManifestWork &&
						managedRBAC.Status.Conditions[0].Status == metav1.ConditionFalse
				}
				return false
			}).Should(BeTrue())

			By("Deleting the ManagedRBAC")
			Expect(k8sClient.Delete(ctx, &managedRBAC)).Should(Succeed())

			By("Creating the ManagedRBAC with invalid RoleBinding ManagedServiceAccount")
			managedRBAC = mrbacv1alpha1.ManagedRBAC{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mbacName,
					Namespace: clusterName,
				},
				Spec: mrbacv1alpha1.ManagedRBACSpec{
					ClusterRole: &mrbacv1alpha1.ClusterRole{
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{"apps"},
							Resources: []string{"services"},
							Verbs:     []string{"create"},
						}},
					},
					RoleBindings: &[]mrbacv1alpha1.RoleBinding{
						{
							Subject: rbacv1.Subject{
								APIGroup: "authentication.open-cluster-management.io",
								Kind:     "ManagedServiceAccount",
								Name:     msaName,
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, &managedRBAC)).Should(Succeed())
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return true
				}
				return false
			}).Should(BeTrue())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mbacKey, &managedRBAC); err != nil {
					return false
				}
				if len(managedRBAC.Status.Conditions) > 0 {
					return managedRBAC.Status.Conditions[0].Type == mrbacv1alpha1.ConditionTypeAppliedRBACManifestWork &&
						managedRBAC.Status.Conditions[0].Status == metav1.ConditionFalse
				}
				return false
			}).Should(BeTrue())

			By("Deleting the ManagedRBAC")
			Expect(k8sClient.Delete(ctx, &managedRBAC)).Should(Succeed())

			By("Creating the ManagedRBAC with invalid managed cluster")
			managedRBAC = mrbacv1alpha1.ManagedRBAC{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mbacName,
					Namespace: "default",
				},
				Spec: mrbacv1alpha1.ManagedRBACSpec{
					ClusterRole: &mrbacv1alpha1.ClusterRole{
						Rules: []rbacv1.PolicyRule{{
							APIGroups: []string{"apps"},
							Resources: []string{"services"},
							Verbs:     []string{"create"},
						}},
					},
				},
			}

			Expect(k8sClient.Create(ctx, &managedRBAC)).Should(Succeed())
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return true
				}
				return false
			}).Should(BeTrue())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: mbacName, Namespace: "default"}, &managedRBAC); err != nil {
					return false
				}
				if len(managedRBAC.Status.Conditions) > 0 {
					return managedRBAC.Status.Conditions[0].Type == mrbacv1alpha1.ConditionTypeValidation &&
						managedRBAC.Status.Conditions[0].Status == metav1.ConditionFalse
				}
				return false
			}).Should(BeTrue())

			By("Deleting the ManagedRBAC")
			Expect(k8sClient.Delete(ctx, &managedRBAC)).Should(Succeed())
		})
	})
})
