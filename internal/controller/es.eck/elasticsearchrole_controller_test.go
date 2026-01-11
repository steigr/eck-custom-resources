/*
Copyright 2022.

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

package eseck

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	eseckv1alpha1 "eck-custom-resources/api/es.eck/v1alpha1"
)

var _ = Describe("ElasticsearchRole Controller", func() {
	const (
		RoleNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an ElasticsearchRole", func() {
		It("Should create the ElasticsearchRole resource successfully", func() {
			ctx := context.Background()

			roleName := "test-role"
			role := &eseckv1alpha1.ElasticsearchRole{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleName,
					Namespace: RoleNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchRoleSpec{
					Body: `{
						"cluster": ["monitor"],
						"indices": [
							{
								"names": ["logs-*"],
								"privileges": ["read"]
							}
						]
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, role)).Should(Succeed())

			roleLookupKey := types.NamespacedName{Name: roleName, Namespace: RoleNamespace}
			createdRole := &eseckv1alpha1.ElasticsearchRole{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, roleLookupKey, createdRole)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdRole.Spec.Body).Should(ContainSubstring("cluster"))
			Expect(createdRole.Spec.Body).Should(ContainSubstring("indices"))
		})

		It("Should create ElasticsearchRole with all privileges", func() {
			ctx := context.Background()

			roleName := "test-role-all-privileges"
			role := &eseckv1alpha1.ElasticsearchRole{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleName,
					Namespace: RoleNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchRoleSpec{
					Body: `{
						"cluster": ["all"],
						"indices": [
							{
								"names": ["*"],
								"privileges": ["all"]
							}
						],
						"applications": [
							{
								"application": "kibana-.kibana",
								"privileges": ["all"],
								"resources": ["*"]
							}
						],
						"run_as": ["other_user"]
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, role)).Should(Succeed())

			roleLookupKey := types.NamespacedName{Name: roleName, Namespace: RoleNamespace}
			createdRole := &eseckv1alpha1.ElasticsearchRole{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, roleLookupKey, createdRole)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdRole.Spec.Body).Should(ContainSubstring("applications"))
			Expect(createdRole.Spec.Body).Should(ContainSubstring("run_as"))
		})

		It("Should create ElasticsearchRole with target instance config", func() {
			ctx := context.Background()

			roleName := "test-role-target"
			role := &eseckv1alpha1.ElasticsearchRole{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleName,
					Namespace: RoleNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchRoleSpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance: "my-elasticsearch",
					},
					Body: `{
						"cluster": ["monitor"],
						"indices": []
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, role)).Should(Succeed())

			roleLookupKey := types.NamespacedName{Name: roleName, Namespace: RoleNamespace}
			createdRole := &eseckv1alpha1.ElasticsearchRole{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, roleLookupKey, createdRole)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdRole.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
		})
	})

	Context("When updating an ElasticsearchRole", func() {
		It("Should update the ElasticsearchRole body successfully", func() {
			ctx := context.Background()

			roleName := "test-role-update"
			role := &eseckv1alpha1.ElasticsearchRole{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleName,
					Namespace: RoleNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchRoleSpec{
					Body: `{"cluster": ["monitor"], "indices": []}`,
				},
			}

			Expect(k8sClient.Create(ctx, role)).Should(Succeed())

			roleLookupKey := types.NamespacedName{Name: roleName, Namespace: RoleNamespace}
			createdRole := &eseckv1alpha1.ElasticsearchRole{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, roleLookupKey, createdRole)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdRole.Spec.Body = `{"cluster": ["all"], "indices": [{"names": ["*"], "privileges": ["read"]}]}`
			Expect(k8sClient.Update(ctx, createdRole)).Should(Succeed())

			updatedRole := &eseckv1alpha1.ElasticsearchRole{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, roleLookupKey, updatedRole)
				if err != nil {
					return false
				}
				return updatedRole.Spec.Body != role.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedRole.Spec.Body).Should(ContainSubstring("all"))
		})
	})

	Context("When deleting an ElasticsearchRole", func() {
		It("Should delete the ElasticsearchRole resource successfully", func() {
			ctx := context.Background()

			roleName := "test-role-delete"
			role := &eseckv1alpha1.ElasticsearchRole{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleName,
					Namespace: RoleNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchRoleSpec{
					Body: `{"cluster": [], "indices": []}`,
				},
			}

			Expect(k8sClient.Create(ctx, role)).Should(Succeed())

			roleLookupKey := types.NamespacedName{Name: roleName, Namespace: RoleNamespace}
			createdRole := &eseckv1alpha1.ElasticsearchRole{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, roleLookupKey, createdRole)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdRole)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, roleLookupKey, &eseckv1alpha1.ElasticsearchRole{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
