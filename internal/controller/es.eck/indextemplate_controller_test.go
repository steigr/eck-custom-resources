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

var _ = Describe("IndexTemplate Controller", func() {
	const (
		IndexTemplateNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an IndexTemplate", func() {
		It("Should create the IndexTemplate resource successfully", func() {
			ctx := context.Background()

			indexTemplateName := "test-index-template"
			indexTemplate := &eseckv1alpha1.IndexTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexTemplateName,
					Namespace: IndexTemplateNamespace,
				},
				Spec: eseckv1alpha1.IndexTemplateSpec{
					Body: `{
						"index_patterns": ["logs-*"],
						"template": {
							"settings": {
								"number_of_shards": 1
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, indexTemplate)).Should(Succeed())

			indexTemplateLookupKey := types.NamespacedName{Name: indexTemplateName, Namespace: IndexTemplateNamespace}
			createdIndexTemplate := &eseckv1alpha1.IndexTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, createdIndexTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndexTemplate.Spec.Body).Should(ContainSubstring("index_patterns"))
			Expect(createdIndexTemplate.Spec.Body).Should(ContainSubstring("logs-*"))
		})

		It("Should create IndexTemplate with composed_of", func() {
			ctx := context.Background()

			indexTemplateName := "test-index-template-composed"
			indexTemplate := &eseckv1alpha1.IndexTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexTemplateName,
					Namespace: IndexTemplateNamespace,
				},
				Spec: eseckv1alpha1.IndexTemplateSpec{
					Body: `{
						"index_patterns": ["metrics-*"],
						"composed_of": ["component1", "component2"],
						"priority": 100,
						"template": {
							"settings": {
								"number_of_shards": 2,
								"number_of_replicas": 1
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, indexTemplate)).Should(Succeed())

			indexTemplateLookupKey := types.NamespacedName{Name: indexTemplateName, Namespace: IndexTemplateNamespace}
			createdIndexTemplate := &eseckv1alpha1.IndexTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, createdIndexTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndexTemplate.Spec.Body).Should(ContainSubstring("composed_of"))
			Expect(createdIndexTemplate.Spec.Body).Should(ContainSubstring("component1"))
		})

		It("Should create IndexTemplate with mappings", func() {
			ctx := context.Background()

			indexTemplateName := "test-index-template-mappings"
			indexTemplate := &eseckv1alpha1.IndexTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexTemplateName,
					Namespace: IndexTemplateNamespace,
				},
				Spec: eseckv1alpha1.IndexTemplateSpec{
					Body: `{
						"index_patterns": ["events-*"],
						"template": {
							"mappings": {
								"properties": {
									"@timestamp": {"type": "date"},
									"message": {"type": "text"},
									"level": {"type": "keyword"}
								}
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, indexTemplate)).Should(Succeed())

			indexTemplateLookupKey := types.NamespacedName{Name: indexTemplateName, Namespace: IndexTemplateNamespace}
			createdIndexTemplate := &eseckv1alpha1.IndexTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, createdIndexTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndexTemplate.Spec.Body).Should(ContainSubstring("mappings"))
			Expect(createdIndexTemplate.Spec.Body).Should(ContainSubstring("@timestamp"))
		})

		It("Should create IndexTemplate with target instance config", func() {
			ctx := context.Background()

			indexTemplateName := "test-index-template-target"
			indexTemplate := &eseckv1alpha1.IndexTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexTemplateName,
					Namespace: IndexTemplateNamespace,
				},
				Spec: eseckv1alpha1.IndexTemplateSpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance:          "my-elasticsearch",
						ElasticsearchInstanceNamespace: "elastic-system",
					},
					Body: `{
						"index_patterns": ["app-*"],
						"template": {}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, indexTemplate)).Should(Succeed())

			indexTemplateLookupKey := types.NamespacedName{Name: indexTemplateName, Namespace: IndexTemplateNamespace}
			createdIndexTemplate := &eseckv1alpha1.IndexTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, createdIndexTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndexTemplate.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
			Expect(createdIndexTemplate.Spec.TargetConfig.ElasticsearchInstanceNamespace).Should(Equal("elastic-system"))
		})
	})

	Context("When updating an IndexTemplate", func() {
		It("Should update the IndexTemplate body successfully", func() {
			ctx := context.Background()

			indexTemplateName := "test-index-template-update"
			indexTemplate := &eseckv1alpha1.IndexTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexTemplateName,
					Namespace: IndexTemplateNamespace,
				},
				Spec: eseckv1alpha1.IndexTemplateSpec{
					Body: `{"index_patterns": ["old-*"], "template": {}}`,
				},
			}

			Expect(k8sClient.Create(ctx, indexTemplate)).Should(Succeed())

			indexTemplateLookupKey := types.NamespacedName{Name: indexTemplateName, Namespace: IndexTemplateNamespace}
			createdIndexTemplate := &eseckv1alpha1.IndexTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, createdIndexTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdIndexTemplate.Spec.Body = `{"index_patterns": ["new-*"], "template": {}}`
			Expect(k8sClient.Update(ctx, createdIndexTemplate)).Should(Succeed())

			updatedIndexTemplate := &eseckv1alpha1.IndexTemplate{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, updatedIndexTemplate)
				if err != nil {
					return false
				}
				return updatedIndexTemplate.Spec.Body != indexTemplate.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedIndexTemplate.Spec.Body).Should(ContainSubstring("new-*"))
		})
	})

	Context("When deleting an IndexTemplate", func() {
		It("Should delete the IndexTemplate resource successfully", func() {
			ctx := context.Background()

			indexTemplateName := "test-index-template-delete"
			indexTemplate := &eseckv1alpha1.IndexTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IndexTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexTemplateName,
					Namespace: IndexTemplateNamespace,
				},
				Spec: eseckv1alpha1.IndexTemplateSpec{
					Body: `{"index_patterns": ["delete-*"], "template": {}}`,
				},
			}

			Expect(k8sClient.Create(ctx, indexTemplate)).Should(Succeed())

			indexTemplateLookupKey := types.NamespacedName{Name: indexTemplateName, Namespace: IndexTemplateNamespace}
			createdIndexTemplate := &eseckv1alpha1.IndexTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, createdIndexTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdIndexTemplate)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexTemplateLookupKey, &eseckv1alpha1.IndexTemplate{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
