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

package kibanaeck

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kibanaeckv1alpha1 "eck-custom-resources/api/kibana.eck/v1alpha1"
)

var _ = Describe("Space Controller", func() {
	const (
		SpaceNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a Space", func() {
		It("Should create the Space resource successfully", func() {
			ctx := context.Background()

			spaceName := "test-space"
			space := &kibanaeckv1alpha1.Space{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Space",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      spaceName,
					Namespace: SpaceNamespace,
				},
				Spec: kibanaeckv1alpha1.SpaceSpec{
					Body: `{
						"name": "Test Space",
						"description": "A test space for unit tests",
						"color": "#aabbcc"
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, space)).Should(Succeed())

			spaceLookupKey := types.NamespacedName{Name: spaceName, Namespace: SpaceNamespace}
			createdSpace := &kibanaeckv1alpha1.Space{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, spaceLookupKey, createdSpace)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSpace.Spec.Body).Should(ContainSubstring("Test Space"))
			Expect(createdSpace.Spec.Body).Should(ContainSubstring("description"))
		})

		It("Should create Space with disabled features", func() {
			ctx := context.Background()

			spaceName := "test-space-features"
			space := &kibanaeckv1alpha1.Space{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Space",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      spaceName,
					Namespace: SpaceNamespace,
				},
				Spec: kibanaeckv1alpha1.SpaceSpec{
					Body: `{
						"name": "Limited Space",
						"description": "Space with limited features",
						"disabledFeatures": ["apm", "ml"]
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, space)).Should(Succeed())

			spaceLookupKey := types.NamespacedName{Name: spaceName, Namespace: SpaceNamespace}
			createdSpace := &kibanaeckv1alpha1.Space{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, spaceLookupKey, createdSpace)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSpace.Spec.Body).Should(ContainSubstring("disabledFeatures"))
			Expect(createdSpace.Spec.Body).Should(ContainSubstring("apm"))
		})

		It("Should create Space with target instance config", func() {
			ctx := context.Background()

			spaceName := "test-space-target"
			space := &kibanaeckv1alpha1.Space{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Space",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      spaceName,
					Namespace: SpaceNamespace,
				},
				Spec: kibanaeckv1alpha1.SpaceSpec{
					TargetConfig: kibanaeckv1alpha1.CommonKibanaConfig{
						KibanaInstance: "my-kibana",
					},
					Body: `{"name": "Space with Target"}`,
				},
			}

			Expect(k8sClient.Create(ctx, space)).Should(Succeed())

			spaceLookupKey := types.NamespacedName{Name: spaceName, Namespace: SpaceNamespace}
			createdSpace := &kibanaeckv1alpha1.Space{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, spaceLookupKey, createdSpace)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSpace.Spec.TargetConfig.KibanaInstance).Should(Equal("my-kibana"))
		})
	})

	Context("When updating a Space", func() {
		It("Should update the Space body successfully", func() {
			ctx := context.Background()

			spaceName := "test-space-update"
			space := &kibanaeckv1alpha1.Space{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Space",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      spaceName,
					Namespace: SpaceNamespace,
				},
				Spec: kibanaeckv1alpha1.SpaceSpec{
					Body: `{"name": "Original Space"}`,
				},
			}

			Expect(k8sClient.Create(ctx, space)).Should(Succeed())

			spaceLookupKey := types.NamespacedName{Name: spaceName, Namespace: SpaceNamespace}
			createdSpace := &kibanaeckv1alpha1.Space{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, spaceLookupKey, createdSpace)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdSpace.Spec.Body = `{"name": "Updated Space", "color": "#ff0000"}`
			Expect(k8sClient.Update(ctx, createdSpace)).Should(Succeed())

			updatedSpace := &kibanaeckv1alpha1.Space{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, spaceLookupKey, updatedSpace)
				if err != nil {
					return false
				}
				return updatedSpace.Spec.Body != space.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedSpace.Spec.Body).Should(ContainSubstring("Updated Space"))
		})
	})

	Context("When deleting a Space", func() {
		It("Should delete the Space resource successfully", func() {
			ctx := context.Background()

			spaceName := "test-space-delete"
			space := &kibanaeckv1alpha1.Space{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Space",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      spaceName,
					Namespace: SpaceNamespace,
				},
				Spec: kibanaeckv1alpha1.SpaceSpec{
					Body: `{"name": "Space to Delete"}`,
				},
			}

			Expect(k8sClient.Create(ctx, space)).Should(Succeed())

			spaceLookupKey := types.NamespacedName{Name: spaceName, Namespace: SpaceNamespace}
			createdSpace := &kibanaeckv1alpha1.Space{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, spaceLookupKey, createdSpace)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdSpace)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, spaceLookupKey, &kibanaeckv1alpha1.Space{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
