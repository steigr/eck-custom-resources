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

var _ = Describe("ElasticsearchUser Controller", func() {
	const (
		UserNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an ElasticsearchUser", func() {
		It("Should create the ElasticsearchUser resource successfully", func() {
			ctx := context.Background()

			userName := "test-user"
			user := &eseckv1alpha1.ElasticsearchUser{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchUser",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      userName,
					Namespace: UserNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchUserSpec{
					Body: `{
						"roles": ["viewer"],
						"full_name": "Test User",
						"email": "test@example.com"
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, user)).Should(Succeed())

			userLookupKey := types.NamespacedName{Name: userName, Namespace: UserNamespace}
			createdUser := &eseckv1alpha1.ElasticsearchUser{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, userLookupKey, createdUser)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdUser.Spec.Body).Should(ContainSubstring("roles"))
			Expect(createdUser.Spec.Body).Should(ContainSubstring("viewer"))
		})

		It("Should create ElasticsearchUser with multiple roles", func() {
			ctx := context.Background()

			userName := "test-user-multi-roles"
			user := &eseckv1alpha1.ElasticsearchUser{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchUser",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      userName,
					Namespace: UserNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchUserSpec{
					Body: `{
						"roles": ["viewer", "editor", "custom_role"],
						"full_name": "Multi Role User",
						"email": "multi@example.com",
						"metadata": {
							"department": "engineering"
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, user)).Should(Succeed())

			userLookupKey := types.NamespacedName{Name: userName, Namespace: UserNamespace}
			createdUser := &eseckv1alpha1.ElasticsearchUser{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, userLookupKey, createdUser)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdUser.Spec.Body).Should(ContainSubstring("viewer"))
			Expect(createdUser.Spec.Body).Should(ContainSubstring("editor"))
			Expect(createdUser.Spec.Body).Should(ContainSubstring("custom_role"))
			Expect(createdUser.Spec.Body).Should(ContainSubstring("metadata"))
		})

		It("Should create ElasticsearchUser with target instance config", func() {
			ctx := context.Background()

			userName := "test-user-target"
			user := &eseckv1alpha1.ElasticsearchUser{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchUser",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      userName,
					Namespace: UserNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchUserSpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance:          "my-elasticsearch",
						ElasticsearchInstanceNamespace: "elastic-system",
					},
					Body: `{
						"roles": ["viewer"]
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, user)).Should(Succeed())

			userLookupKey := types.NamespacedName{Name: userName, Namespace: UserNamespace}
			createdUser := &eseckv1alpha1.ElasticsearchUser{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, userLookupKey, createdUser)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdUser.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
			Expect(createdUser.Spec.TargetConfig.ElasticsearchInstanceNamespace).Should(Equal("elastic-system"))
		})
	})

	Context("When updating an ElasticsearchUser", func() {
		It("Should update the ElasticsearchUser body successfully", func() {
			ctx := context.Background()

			userName := "test-user-update"
			user := &eseckv1alpha1.ElasticsearchUser{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchUser",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      userName,
					Namespace: UserNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchUserSpec{
					Body: `{"roles": ["viewer"], "full_name": "Original Name"}`,
				},
			}

			Expect(k8sClient.Create(ctx, user)).Should(Succeed())

			userLookupKey := types.NamespacedName{Name: userName, Namespace: UserNamespace}
			createdUser := &eseckv1alpha1.ElasticsearchUser{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, userLookupKey, createdUser)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdUser.Spec.Body = `{"roles": ["editor"], "full_name": "Updated Name"}`
			Expect(k8sClient.Update(ctx, createdUser)).Should(Succeed())

			updatedUser := &eseckv1alpha1.ElasticsearchUser{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, userLookupKey, updatedUser)
				if err != nil {
					return false
				}
				return updatedUser.Spec.Body != user.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedUser.Spec.Body).Should(ContainSubstring("editor"))
			Expect(updatedUser.Spec.Body).Should(ContainSubstring("Updated Name"))
		})
	})

	Context("When deleting an ElasticsearchUser", func() {
		It("Should delete the ElasticsearchUser resource successfully", func() {
			ctx := context.Background()

			userName := "test-user-delete"
			user := &eseckv1alpha1.ElasticsearchUser{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ElasticsearchUser",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      userName,
					Namespace: UserNamespace,
				},
				Spec: eseckv1alpha1.ElasticsearchUserSpec{
					Body: `{"roles": []}`,
				},
			}

			Expect(k8sClient.Create(ctx, user)).Should(Succeed())

			userLookupKey := types.NamespacedName{Name: userName, Namespace: UserNamespace}
			createdUser := &eseckv1alpha1.ElasticsearchUser{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, userLookupKey, createdUser)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdUser)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, userLookupKey, &eseckv1alpha1.ElasticsearchUser{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
