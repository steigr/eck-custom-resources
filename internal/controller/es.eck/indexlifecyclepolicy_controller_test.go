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

var _ = Describe("IndexLifecyclePolicy Controller", func() {
	const (
		ILPNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an IndexLifecyclePolicy", func() {
		It("Should create the IndexLifecyclePolicy resource successfully", func() {
			ctx := context.Background()

			ilpName := "test-ilp"
			ilp := &eseckv1alpha1.IndexLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ilpName,
					Namespace: ILPNamespace,
				},
				Spec: eseckv1alpha1.IndexLifecyclePolicySpec{
					Body: `{
						"policy": {
							"phases": {
								"hot": {
									"actions": {
										"rollover": {
											"max_age": "7d",
											"max_size": "50gb"
										}
									}
								}
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, ilp)).Should(Succeed())

			ilpLookupKey := types.NamespacedName{Name: ilpName, Namespace: ILPNamespace}
			createdILP := &eseckv1alpha1.IndexLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ilpLookupKey, createdILP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdILP.Spec.Body).Should(ContainSubstring("policy"))
			Expect(createdILP.Spec.Body).Should(ContainSubstring("phases"))
			Expect(createdILP.Spec.Body).Should(ContainSubstring("hot"))
		})

		It("Should create IndexLifecyclePolicy with multiple phases", func() {
			ctx := context.Background()

			ilpName := "test-ilp-phases"
			ilp := &eseckv1alpha1.IndexLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ilpName,
					Namespace: ILPNamespace,
				},
				Spec: eseckv1alpha1.IndexLifecyclePolicySpec{
					Body: `{
						"policy": {
							"phases": {
								"hot": {
									"min_age": "0ms",
									"actions": {
										"rollover": {
											"max_age": "1d"
										}
									}
								},
								"warm": {
									"min_age": "7d",
									"actions": {
										"shrink": {
											"number_of_shards": 1
										}
									}
								},
								"cold": {
									"min_age": "30d",
									"actions": {
										"freeze": {}
									}
								},
								"delete": {
									"min_age": "90d",
									"actions": {
										"delete": {}
									}
								}
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, ilp)).Should(Succeed())

			ilpLookupKey := types.NamespacedName{Name: ilpName, Namespace: ILPNamespace}
			createdILP := &eseckv1alpha1.IndexLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ilpLookupKey, createdILP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdILP.Spec.Body).Should(ContainSubstring("hot"))
			Expect(createdILP.Spec.Body).Should(ContainSubstring("warm"))
			Expect(createdILP.Spec.Body).Should(ContainSubstring("cold"))
			Expect(createdILP.Spec.Body).Should(ContainSubstring("delete"))
		})

		It("Should create IndexLifecyclePolicy with target instance config", func() {
			ctx := context.Background()

			ilpName := "test-ilp-target"
			ilp := &eseckv1alpha1.IndexLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ilpName,
					Namespace: ILPNamespace,
				},
				Spec: eseckv1alpha1.IndexLifecyclePolicySpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance: "my-elasticsearch",
					},
					Body: `{
						"policy": {
							"phases": {
								"hot": {
									"actions": {}
								}
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, ilp)).Should(Succeed())

			ilpLookupKey := types.NamespacedName{Name: ilpName, Namespace: ILPNamespace}
			createdILP := &eseckv1alpha1.IndexLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ilpLookupKey, createdILP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdILP.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
		})
	})

	Context("When updating an IndexLifecyclePolicy", func() {
		It("Should update the IndexLifecyclePolicy body successfully", func() {
			ctx := context.Background()

			ilpName := "test-ilp-update"
			ilp := &eseckv1alpha1.IndexLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ilpName,
					Namespace: ILPNamespace,
				},
				Spec: eseckv1alpha1.IndexLifecyclePolicySpec{
					Body: `{"policy": {"phases": {"hot": {"actions": {"rollover": {"max_age": "1d"}}}}}}`,
				},
			}

			Expect(k8sClient.Create(ctx, ilp)).Should(Succeed())

			ilpLookupKey := types.NamespacedName{Name: ilpName, Namespace: ILPNamespace}
			createdILP := &eseckv1alpha1.IndexLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ilpLookupKey, createdILP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdILP.Spec.Body = `{"policy": {"phases": {"hot": {"actions": {"rollover": {"max_age": "7d"}}}}}}`
			Expect(k8sClient.Update(ctx, createdILP)).Should(Succeed())

			updatedILP := &eseckv1alpha1.IndexLifecyclePolicy{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, ilpLookupKey, updatedILP)
				if err != nil {
					return false
				}
				return updatedILP.Spec.Body != ilp.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedILP.Spec.Body).Should(ContainSubstring("7d"))
		})
	})

	Context("When deleting an IndexLifecyclePolicy", func() {
		It("Should delete the IndexLifecyclePolicy resource successfully", func() {
			ctx := context.Background()

			ilpName := "test-ilp-delete"
			ilp := &eseckv1alpha1.IndexLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ilpName,
					Namespace: ILPNamespace,
				},
				Spec: eseckv1alpha1.IndexLifecyclePolicySpec{
					Body: `{"policy": {"phases": {}}}`,
				},
			}

			Expect(k8sClient.Create(ctx, ilp)).Should(Succeed())

			ilpLookupKey := types.NamespacedName{Name: ilpName, Namespace: ILPNamespace}
			createdILP := &eseckv1alpha1.IndexLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ilpLookupKey, createdILP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdILP)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, ilpLookupKey, &eseckv1alpha1.IndexLifecyclePolicy{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
