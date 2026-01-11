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

var _ = Describe("Visualization Controller", func() {
	const (
		VizNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a Visualization", func() {
		It("Should create the Visualization resource successfully", func() {
			ctx := context.Background()

			vizName := "test-visualization"
			viz := &kibanaeckv1alpha1.Visualization{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Visualization",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      vizName,
					Namespace: VizNamespace,
				},
				Spec: kibanaeckv1alpha1.VisualizationSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{
							"title": "Test Visualization",
							"visState": "{}",
							"uiStateJSON": "{}"
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, viz)).Should(Succeed())

			vizLookupKey := types.NamespacedName{Name: vizName, Namespace: VizNamespace}
			createdViz := &kibanaeckv1alpha1.Visualization{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, vizLookupKey, createdViz)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdViz.Spec.Body).Should(ContainSubstring("Test Visualization"))
		})

		It("Should create Visualization with space", func() {
			ctx := context.Background()

			vizName := "test-visualization-space"
			space := "analytics"
			viz := &kibanaeckv1alpha1.Visualization{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Visualization",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      vizName,
					Namespace: VizNamespace,
				},
				Spec: kibanaeckv1alpha1.VisualizationSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Space: &space,
						Body: `{
							"title": "Viz in Space",
							"visState": "{}"
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, viz)).Should(Succeed())

			vizLookupKey := types.NamespacedName{Name: vizName, Namespace: VizNamespace}
			createdViz := &kibanaeckv1alpha1.Visualization{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, vizLookupKey, createdViz)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdViz.Spec.Space).ShouldNot(BeNil())
			Expect(*createdViz.Spec.Space).Should(Equal("analytics"))
		})

		It("Should create Visualization with target instance config", func() {
			ctx := context.Background()

			vizName := "test-visualization-target"
			viz := &kibanaeckv1alpha1.Visualization{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Visualization",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      vizName,
					Namespace: VizNamespace,
				},
				Spec: kibanaeckv1alpha1.VisualizationSpec{
					TargetConfig: kibanaeckv1alpha1.CommonKibanaConfig{
						KibanaInstance: "my-kibana",
					},
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Viz with Target"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, viz)).Should(Succeed())

			vizLookupKey := types.NamespacedName{Name: vizName, Namespace: VizNamespace}
			createdViz := &kibanaeckv1alpha1.Visualization{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, vizLookupKey, createdViz)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdViz.Spec.TargetConfig.KibanaInstance).Should(Equal("my-kibana"))
		})
	})

	Context("When deleting a Visualization", func() {
		It("Should delete the Visualization resource successfully", func() {
			ctx := context.Background()

			vizName := "test-visualization-delete"
			viz := &kibanaeckv1alpha1.Visualization{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Visualization",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      vizName,
					Namespace: VizNamespace,
				},
				Spec: kibanaeckv1alpha1.VisualizationSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Viz to Delete"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, viz)).Should(Succeed())

			vizLookupKey := types.NamespacedName{Name: vizName, Namespace: VizNamespace}
			createdViz := &kibanaeckv1alpha1.Visualization{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, vizLookupKey, createdViz)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdViz)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, vizLookupKey, &kibanaeckv1alpha1.Visualization{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = Describe("Lens Controller", func() {
	const (
		LensNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a Lens", func() {
		It("Should create the Lens resource successfully", func() {
			ctx := context.Background()

			lensName := "test-lens"
			lens := &kibanaeckv1alpha1.Lens{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Lens",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      lensName,
					Namespace: LensNamespace,
				},
				Spec: kibanaeckv1alpha1.LensSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{
							"title": "Test Lens",
							"visualizationType": "lnsXY",
							"state": {}
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, lens)).Should(Succeed())

			lensLookupKey := types.NamespacedName{Name: lensName, Namespace: LensNamespace}
			createdLens := &kibanaeckv1alpha1.Lens{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lensLookupKey, createdLens)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdLens.Spec.Body).Should(ContainSubstring("Test Lens"))
			Expect(createdLens.Spec.Body).Should(ContainSubstring("lnsXY"))
		})

		It("Should create Lens with target instance config", func() {
			ctx := context.Background()

			lensName := "test-lens-target"
			lens := &kibanaeckv1alpha1.Lens{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Lens",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      lensName,
					Namespace: LensNamespace,
				},
				Spec: kibanaeckv1alpha1.LensSpec{
					TargetConfig: kibanaeckv1alpha1.CommonKibanaConfig{
						KibanaInstance: "my-kibana",
					},
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Lens with Target"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, lens)).Should(Succeed())

			lensLookupKey := types.NamespacedName{Name: lensName, Namespace: LensNamespace}
			createdLens := &kibanaeckv1alpha1.Lens{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lensLookupKey, createdLens)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdLens.Spec.TargetConfig.KibanaInstance).Should(Equal("my-kibana"))
		})
	})

	Context("When deleting a Lens", func() {
		It("Should delete the Lens resource successfully", func() {
			ctx := context.Background()

			lensName := "test-lens-delete"
			lens := &kibanaeckv1alpha1.Lens{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Lens",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      lensName,
					Namespace: LensNamespace,
				},
				Spec: kibanaeckv1alpha1.LensSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Lens to Delete"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, lens)).Should(Succeed())

			lensLookupKey := types.NamespacedName{Name: lensName, Namespace: LensNamespace}
			createdLens := &kibanaeckv1alpha1.Lens{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lensLookupKey, createdLens)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdLens)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lensLookupKey, &kibanaeckv1alpha1.Lens{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = Describe("SavedSearch Controller", func() {
	const (
		SearchNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a SavedSearch", func() {
		It("Should create the SavedSearch resource successfully", func() {
			ctx := context.Background()

			searchName := "test-saved-search"
			search := &kibanaeckv1alpha1.SavedSearch{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "SavedSearch",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      searchName,
					Namespace: SearchNamespace,
				},
				Spec: kibanaeckv1alpha1.SavedSearchSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{
							"title": "Test Search",
							"columns": ["_source"],
							"sort": []
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, search)).Should(Succeed())

			searchLookupKey := types.NamespacedName{Name: searchName, Namespace: SearchNamespace}
			createdSearch := &kibanaeckv1alpha1.SavedSearch{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, searchLookupKey, createdSearch)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSearch.Spec.Body).Should(ContainSubstring("Test Search"))
		})

		It("Should create SavedSearch with target instance config", func() {
			ctx := context.Background()

			searchName := "test-saved-search-target"
			search := &kibanaeckv1alpha1.SavedSearch{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "SavedSearch",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      searchName,
					Namespace: SearchNamespace,
				},
				Spec: kibanaeckv1alpha1.SavedSearchSpec{
					TargetConfig: kibanaeckv1alpha1.CommonKibanaConfig{
						KibanaInstance: "my-kibana",
					},
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Search with Target"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, search)).Should(Succeed())

			searchLookupKey := types.NamespacedName{Name: searchName, Namespace: SearchNamespace}
			createdSearch := &kibanaeckv1alpha1.SavedSearch{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, searchLookupKey, createdSearch)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSearch.Spec.TargetConfig.KibanaInstance).Should(Equal("my-kibana"))
		})
	})

	Context("When deleting a SavedSearch", func() {
		It("Should delete the SavedSearch resource successfully", func() {
			ctx := context.Background()

			searchName := "test-saved-search-delete"
			search := &kibanaeckv1alpha1.SavedSearch{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "SavedSearch",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      searchName,
					Namespace: SearchNamespace,
				},
				Spec: kibanaeckv1alpha1.SavedSearchSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Search to Delete"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, search)).Should(Succeed())

			searchLookupKey := types.NamespacedName{Name: searchName, Namespace: SearchNamespace}
			createdSearch := &kibanaeckv1alpha1.SavedSearch{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, searchLookupKey, createdSearch)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdSearch)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, searchLookupKey, &kibanaeckv1alpha1.SavedSearch{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = Describe("IndexPattern Controller", func() {
	const (
		IndexPatternNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an IndexPattern", func() {
		It("Should create the IndexPattern resource successfully", func() {
			ctx := context.Background()

			indexPatternName := "test-index-pattern"
			indexPattern := &kibanaeckv1alpha1.IndexPattern{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "IndexPattern",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexPatternName,
					Namespace: IndexPatternNamespace,
				},
				Spec: kibanaeckv1alpha1.IndexPatternSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{
							"title": "logs-*",
							"timeFieldName": "@timestamp"
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, indexPattern)).Should(Succeed())

			indexPatternLookupKey := types.NamespacedName{Name: indexPatternName, Namespace: IndexPatternNamespace}
			createdIndexPattern := &kibanaeckv1alpha1.IndexPattern{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexPatternLookupKey, createdIndexPattern)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndexPattern.Spec.Body).Should(ContainSubstring("logs-*"))
		})

		It("Should create IndexPattern with target instance config", func() {
			ctx := context.Background()

			indexPatternName := "test-index-pattern-target"
			indexPattern := &kibanaeckv1alpha1.IndexPattern{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "IndexPattern",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexPatternName,
					Namespace: IndexPatternNamespace,
				},
				Spec: kibanaeckv1alpha1.IndexPatternSpec{
					TargetConfig: kibanaeckv1alpha1.CommonKibanaConfig{
						KibanaInstance: "my-kibana",
					},
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "metrics-*"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, indexPattern)).Should(Succeed())

			indexPatternLookupKey := types.NamespacedName{Name: indexPatternName, Namespace: IndexPatternNamespace}
			createdIndexPattern := &kibanaeckv1alpha1.IndexPattern{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexPatternLookupKey, createdIndexPattern)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdIndexPattern.Spec.TargetConfig.KibanaInstance).Should(Equal("my-kibana"))
		})
	})

	Context("When deleting an IndexPattern", func() {
		It("Should delete the IndexPattern resource successfully", func() {
			ctx := context.Background()

			indexPatternName := "test-index-pattern-delete"
			indexPattern := &kibanaeckv1alpha1.IndexPattern{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "IndexPattern",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      indexPatternName,
					Namespace: IndexPatternNamespace,
				},
				Spec: kibanaeckv1alpha1.IndexPatternSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "delete-*"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, indexPattern)).Should(Succeed())

			indexPatternLookupKey := types.NamespacedName{Name: indexPatternName, Namespace: IndexPatternNamespace}
			createdIndexPattern := &kibanaeckv1alpha1.IndexPattern{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexPatternLookupKey, createdIndexPattern)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdIndexPattern)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, indexPatternLookupKey, &kibanaeckv1alpha1.IndexPattern{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
