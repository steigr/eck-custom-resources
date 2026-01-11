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

var _ = Describe("Index Controller", func() {
	const (
		IndexNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an Index", func() {
		It("Should create the Index resource successfully", func() {
			ctx := context.Background()

			indexName := "test-index"
			index := &eseckv1alpha1.Index{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "Index",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexName,
					Namespace: IndexNamespace,
				},
				Spec: eseckv1alpha1.IndexSpec{
					Body: `{
						"settings": {
							"number_of_shards": 1,
							"number_of_replicas": 0
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, index)).Should(Succeed())

			indexLookupKey := types.NamespacedName{Name: indexName, Namespace: IndexNamespace}
			createdIndex := &eseckv1alpha1.Index{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexLookupKey, createdIndex)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndex.Spec.Body).Should(ContainSubstring("settings"))
			Expect(createdIndex.Spec.Body).Should(ContainSubstring("number_of_shards"))
		})

		It("Should create Index with mappings", func() {
			ctx := context.Background()

			indexName := "test-index-mappings"
			index := &eseckv1alpha1.Index{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "Index",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexName,
					Namespace: IndexNamespace,
				},
				Spec: eseckv1alpha1.IndexSpec{
					Body: `{
						"settings": {
							"number_of_shards": 1
						},
						"mappings": {
							"properties": {
								"@timestamp": {"type": "date"},
								"message": {"type": "text"},
								"level": {"type": "keyword"},
								"source": {"type": "keyword"}
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, index)).Should(Succeed())

			indexLookupKey := types.NamespacedName{Name: indexName, Namespace: IndexNamespace}
			createdIndex := &eseckv1alpha1.Index{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexLookupKey, createdIndex)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndex.Spec.Body).Should(ContainSubstring("mappings"))
			Expect(createdIndex.Spec.Body).Should(ContainSubstring("@timestamp"))
			Expect(createdIndex.Spec.Body).Should(ContainSubstring("message"))
		})

		It("Should create Index with aliases", func() {
			ctx := context.Background()

			indexName := "test-index-aliases"
			index := &eseckv1alpha1.Index{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "Index",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexName,
					Namespace: IndexNamespace,
				},
				Spec: eseckv1alpha1.IndexSpec{
					Body: `{
						"settings": {
							"number_of_shards": 1
						},
						"aliases": {
							"logs": {},
							"logs-current": {
								"is_write_index": true
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, index)).Should(Succeed())

			indexLookupKey := types.NamespacedName{Name: indexName, Namespace: IndexNamespace}
			createdIndex := &eseckv1alpha1.Index{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexLookupKey, createdIndex)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndex.Spec.Body).Should(ContainSubstring("aliases"))
			Expect(createdIndex.Spec.Body).Should(ContainSubstring("logs"))
		})

		It("Should create Index with target instance config", func() {
			ctx := context.Background()

			indexName := "test-index-target"
			index := &eseckv1alpha1.Index{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "Index",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexName,
					Namespace: IndexNamespace,
				},
				Spec: eseckv1alpha1.IndexSpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance:          "my-elasticsearch",
						ElasticsearchInstanceNamespace: "elastic-system",
					},
					Body: `{
						"settings": {
							"number_of_shards": 1
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, index)).Should(Succeed())

			indexLookupKey := types.NamespacedName{Name: indexName, Namespace: IndexNamespace}
			createdIndex := &eseckv1alpha1.Index{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexLookupKey, createdIndex)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndex.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
			Expect(createdIndex.Spec.TargetConfig.ElasticsearchInstanceNamespace).Should(Equal("elastic-system"))
		})
	})

	Context("When deleting an Index", func() {
		It("Should delete the Index resource successfully", func() {
			ctx := context.Background()

			indexName := "test-index-delete"
			index := &eseckv1alpha1.Index{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "Index",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexName,
					Namespace: IndexNamespace,
				},
				Spec: eseckv1alpha1.IndexSpec{
					Body: `{"settings": {"number_of_shards": 1}}`,
				},
			}

			Expect(k8sClient.Create(ctx, index)).Should(Succeed())

			indexLookupKey := types.NamespacedName{Name: indexName, Namespace: IndexNamespace}
			createdIndex := &eseckv1alpha1.Index{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexLookupKey, createdIndex)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdIndex)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexLookupKey, &eseckv1alpha1.Index{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
