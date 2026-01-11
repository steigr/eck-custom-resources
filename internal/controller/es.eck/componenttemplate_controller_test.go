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

var _ = Describe("ComponentTemplate Controller", func() {
	const (
		ComponentTemplateNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a ComponentTemplate", func() {
		It("Should create the ComponentTemplate resource successfully", func() {
			ctx := context.Background()

			componentTemplateName := "test-component-template"
			componentTemplate := &eseckv1alpha1.ComponentTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ComponentTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      componentTemplateName,
					Namespace: ComponentTemplateNamespace,
				},
				Spec: eseckv1alpha1.ComponentTemplateSpec{
					Body: `{
						"template": {
							"mappings": {
								"properties": {
									"@timestamp": {
										"type": "date"
									}
								}
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, componentTemplate)).Should(Succeed())

			componentTemplateLookupKey := types.NamespacedName{Name: componentTemplateName, Namespace: ComponentTemplateNamespace}
			createdComponentTemplate := &eseckv1alpha1.ComponentTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, componentTemplateLookupKey, createdComponentTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdComponentTemplate.Spec.Body).Should(ContainSubstring("template"))
			Expect(createdComponentTemplate.Spec.Body).Should(ContainSubstring("mappings"))
		})

		It("Should create ComponentTemplate with settings", func() {
			ctx := context.Background()

			componentTemplateName := "test-component-template-settings"
			componentTemplate := &eseckv1alpha1.ComponentTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ComponentTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      componentTemplateName,
					Namespace: ComponentTemplateNamespace,
				},
				Spec: eseckv1alpha1.ComponentTemplateSpec{
					Body: `{
						"template": {
							"settings": {
								"number_of_shards": 1,
								"number_of_replicas": 0
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, componentTemplate)).Should(Succeed())

			componentTemplateLookupKey := types.NamespacedName{Name: componentTemplateName, Namespace: ComponentTemplateNamespace}
			createdComponentTemplate := &eseckv1alpha1.ComponentTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, componentTemplateLookupKey, createdComponentTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdComponentTemplate.Spec.Body).Should(ContainSubstring("number_of_shards"))
		})

		It("Should create ComponentTemplate with target instance config", func() {
			ctx := context.Background()

			componentTemplateName := "test-component-template-target"
			componentTemplate := &eseckv1alpha1.ComponentTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ComponentTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      componentTemplateName,
					Namespace: ComponentTemplateNamespace,
				},
				Spec: eseckv1alpha1.ComponentTemplateSpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance: "my-elasticsearch",
					},
					Body: `{
						"template": {
							"mappings": {
								"properties": {
									"field1": {"type": "keyword"}
								}
							}
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, componentTemplate)).Should(Succeed())

			componentTemplateLookupKey := types.NamespacedName{Name: componentTemplateName, Namespace: ComponentTemplateNamespace}
			createdComponentTemplate := &eseckv1alpha1.ComponentTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, componentTemplateLookupKey, createdComponentTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdComponentTemplate.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
		})
	})

	Context("When updating a ComponentTemplate", func() {
		It("Should update the ComponentTemplate body successfully", func() {
			ctx := context.Background()

			componentTemplateName := "test-component-template-update"
			componentTemplate := &eseckv1alpha1.ComponentTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ComponentTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      componentTemplateName,
					Namespace: ComponentTemplateNamespace,
				},
				Spec: eseckv1alpha1.ComponentTemplateSpec{
					Body: `{"template": {"mappings": {"properties": {"field1": {"type": "text"}}}}}`,
				},
			}

			Expect(k8sClient.Create(ctx, componentTemplate)).Should(Succeed())

			componentTemplateLookupKey := types.NamespacedName{Name: componentTemplateName, Namespace: ComponentTemplateNamespace}
			createdComponentTemplate := &eseckv1alpha1.ComponentTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, componentTemplateLookupKey, createdComponentTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdComponentTemplate.Spec.Body = `{"template": {"mappings": {"properties": {"field1": {"type": "keyword"}}}}}`
			Expect(k8sClient.Update(ctx, createdComponentTemplate)).Should(Succeed())

			updatedComponentTemplate := &eseckv1alpha1.ComponentTemplate{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, componentTemplateLookupKey, updatedComponentTemplate)
				if err != nil {
					return false
				}
				return updatedComponentTemplate.Spec.Body != componentTemplate.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedComponentTemplate.Spec.Body).Should(ContainSubstring("keyword"))
		})
	})

	Context("When deleting a ComponentTemplate", func() {
		It("Should delete the ComponentTemplate resource successfully", func() {
			ctx := context.Background()

			componentTemplateName := "test-component-template-delete"
			componentTemplate := &eseckv1alpha1.ComponentTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "ComponentTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      componentTemplateName,
					Namespace: ComponentTemplateNamespace,
				},
				Spec: eseckv1alpha1.ComponentTemplateSpec{
					Body: `{"template": {"mappings": {}}}`,
				},
			}

			Expect(k8sClient.Create(ctx, componentTemplate)).Should(Succeed())

			componentTemplateLookupKey := types.NamespacedName{Name: componentTemplateName, Namespace: ComponentTemplateNamespace}
			createdComponentTemplate := &eseckv1alpha1.ComponentTemplate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, componentTemplateLookupKey, createdComponentTemplate)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdComponentTemplate)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, componentTemplateLookupKey, &eseckv1alpha1.ComponentTemplate{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
