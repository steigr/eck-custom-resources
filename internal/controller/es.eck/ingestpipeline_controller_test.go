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

var _ = Describe("IngestPipeline Controller", func() {
	const (
		IngestPipelineName      = "test-ingest-pipeline"
		IngestPipelineNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an IngestPipeline", func() {
		It("Should create the IngestPipeline resource successfully", func() {
			ctx := context.Background()

			ingestPipeline := &eseckv1alpha1.IngestPipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IngestPipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      IngestPipelineName,
					Namespace: IngestPipelineNamespace,
				},
				Spec: eseckv1alpha1.IngestPipelineSpec{
					Body: `{
						"description": "Test pipeline",
						"processors": []
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, ingestPipeline)).Should(Succeed())

			ingestPipelineLookupKey := types.NamespacedName{Name: IngestPipelineName, Namespace: IngestPipelineNamespace}
			createdIngestPipeline := &eseckv1alpha1.IngestPipeline{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, createdIngestPipeline)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIngestPipeline.Spec.Body).Should(ContainSubstring("Test pipeline"))
		})

		It("Should have default UpdatePolicy set to Overwrite", func() {
			ctx := context.Background()

			ingestPipelineName := "test-ingest-pipeline-default-policy"
			ingestPipeline := &eseckv1alpha1.IngestPipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IngestPipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingestPipelineName,
					Namespace: IngestPipelineNamespace,
				},
				Spec: eseckv1alpha1.IngestPipelineSpec{
					Body: `{"description": "Test pipeline with default policy", "processors": []}`,
				},
			}

			Expect(k8sClient.Create(ctx, ingestPipeline)).Should(Succeed())

			ingestPipelineLookupKey := types.NamespacedName{Name: ingestPipelineName, Namespace: IngestPipelineNamespace}
			createdIngestPipeline := &eseckv1alpha1.IngestPipeline{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, createdIngestPipeline)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Default UpdateMode should be Overwrite (or empty which is treated as Overwrite)
			Expect(createdIngestPipeline.Spec.UpdatePolicy.UpdateMode).Should(
				SatisfyAny(
					Equal(eseckv1alpha1.UpdateModeOverwrite),
					Equal(eseckv1alpha1.UpdateMode("")),
				),
			)
		})

		It("Should accept UpdatePolicy with Block mode", func() {
			ctx := context.Background()

			ingestPipelineName := "test-ingest-pipeline-block-policy"
			ingestPipeline := &eseckv1alpha1.IngestPipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IngestPipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingestPipelineName,
					Namespace: IngestPipelineNamespace,
				},
				Spec: eseckv1alpha1.IngestPipelineSpec{
					Body: `{"description": "Test pipeline with block policy", "processors": []}`,
					UpdatePolicy: eseckv1alpha1.UpdatePolicySpec{
						UpdateMode: eseckv1alpha1.UpdateModeBlock,
					},
				},
			}

			Expect(k8sClient.Create(ctx, ingestPipeline)).Should(Succeed())

			ingestPipelineLookupKey := types.NamespacedName{Name: ingestPipelineName, Namespace: IngestPipelineNamespace}
			createdIngestPipeline := &eseckv1alpha1.IngestPipeline{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, createdIngestPipeline)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIngestPipeline.Spec.UpdatePolicy.UpdateMode).Should(Equal(eseckv1alpha1.UpdateModeBlock))
		})
	})

	Context("When creating an IngestPipeline with template references", func() {
		It("Should accept IngestPipeline with template spec", func() {
			ctx := context.Background()

			ingestPipelineName := "test-ingest-pipeline-with-template"
			ingestPipeline := &eseckv1alpha1.IngestPipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IngestPipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingestPipelineName,
					Namespace: IngestPipelineNamespace,
				},
				Spec: eseckv1alpha1.IngestPipelineSpec{
					Body: `{"description": "{{ .Values.default.mydata.description }}", "processors": []}`,
					Template: eseckv1alpha1.CommonTemplatingSpec{
						References: []eseckv1alpha1.CommonTemplatingSpecReference{
							{
								Name:      "mydata",
								Namespace: "default",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, ingestPipeline)).Should(Succeed())

			ingestPipelineLookupKey := types.NamespacedName{Name: ingestPipelineName, Namespace: IngestPipelineNamespace}
			createdIngestPipeline := &eseckv1alpha1.IngestPipeline{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, createdIngestPipeline)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIngestPipeline.Spec.Template.References).Should(HaveLen(1))
			Expect(createdIngestPipeline.Spec.Template.References[0].Name).Should(Equal("mydata"))
		})
	})

	Context("When updating an IngestPipeline", func() {
		It("Should update the IngestPipeline body", func() {
			ctx := context.Background()

			ingestPipelineName := "test-ingest-pipeline-update"
			ingestPipeline := &eseckv1alpha1.IngestPipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IngestPipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingestPipelineName,
					Namespace: IngestPipelineNamespace,
				},
				Spec: eseckv1alpha1.IngestPipelineSpec{
					Body: `{"description": "Original description", "processors": []}`,
				},
			}

			Expect(k8sClient.Create(ctx, ingestPipeline)).Should(Succeed())

			ingestPipelineLookupKey := types.NamespacedName{Name: ingestPipelineName, Namespace: IngestPipelineNamespace}
			createdIngestPipeline := &eseckv1alpha1.IngestPipeline{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, createdIngestPipeline)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdIngestPipeline.Spec.Body = `{"description": "Updated description", "processors": []}`
			Expect(k8sClient.Update(ctx, createdIngestPipeline)).Should(Succeed())

			// Verify the update
			updatedIngestPipeline := &eseckv1alpha1.IngestPipeline{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, updatedIngestPipeline)
				if err != nil {
					return false
				}
				return updatedIngestPipeline.Spec.Body == `{"description": "Updated description", "processors": []}`
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting an IngestPipeline", func() {
		It("Should delete the IngestPipeline resource", func() {
			ctx := context.Background()

			ingestPipelineName := "test-ingest-pipeline-delete"
			ingestPipeline := &eseckv1alpha1.IngestPipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "IngestPipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingestPipelineName,
					Namespace: IngestPipelineNamespace,
				},
				Spec: eseckv1alpha1.IngestPipelineSpec{
					Body: `{"description": "Pipeline to delete", "processors": []}`,
				},
			}

			Expect(k8sClient.Create(ctx, ingestPipeline)).Should(Succeed())

			ingestPipelineLookupKey := types.NamespacedName{Name: ingestPipelineName, Namespace: IngestPipelineNamespace}
			createdIngestPipeline := &eseckv1alpha1.IngestPipeline{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, createdIngestPipeline)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the IngestPipeline
			Expect(k8sClient.Delete(ctx, createdIngestPipeline)).Should(Succeed())

			// Verify it's deleted (eventually)
			deletedIngestPipeline := &eseckv1alpha1.IngestPipeline{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, ingestPipelineLookupKey, deletedIngestPipeline)
				return err != nil // Should return error when not found
			}, timeout, interval).Should(BeTrue())
		})
	})
})
