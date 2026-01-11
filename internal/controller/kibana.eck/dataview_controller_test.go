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

var _ = Describe("DataView Controller", func() {
	const (
		DataViewNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a DataView", func() {
		It("Should create the DataView resource successfully", func() {
			ctx := context.Background()

			dataViewName := "test-dataview"
			dataView := &kibanaeckv1alpha1.DataView{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "DataView",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataViewName,
					Namespace: DataViewNamespace,
				},
				Spec: kibanaeckv1alpha1.DataViewSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{
							"title": "logs-*",
							"timeFieldName": "@timestamp"
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dataView)).Should(Succeed())

			dataViewLookupKey := types.NamespacedName{Name: dataViewName, Namespace: DataViewNamespace}
			createdDataView := &kibanaeckv1alpha1.DataView{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dataViewLookupKey, createdDataView)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdDataView.Spec.Body).Should(ContainSubstring("logs-*"))
			Expect(createdDataView.Spec.Body).Should(ContainSubstring("@timestamp"))
		})

		It("Should create DataView with space", func() {
			ctx := context.Background()

			dataViewName := "test-dataview-space"
			space := "analytics"
			dataView := &kibanaeckv1alpha1.DataView{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "DataView",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataViewName,
					Namespace: DataViewNamespace,
				},
				Spec: kibanaeckv1alpha1.DataViewSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Space: &space,
						Body: `{
							"title": "metrics-*",
							"timeFieldName": "timestamp"
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dataView)).Should(Succeed())

			dataViewLookupKey := types.NamespacedName{Name: dataViewName, Namespace: DataViewNamespace}
			createdDataView := &kibanaeckv1alpha1.DataView{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dataViewLookupKey, createdDataView)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdDataView.Spec.Space).ShouldNot(BeNil())
			Expect(*createdDataView.Spec.Space).Should(Equal("analytics"))
		})

		It("Should create DataView with target instance config", func() {
			ctx := context.Background()

			dataViewName := "test-dataview-target"
			dataView := &kibanaeckv1alpha1.DataView{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "DataView",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataViewName,
					Namespace: DataViewNamespace,
				},
				Spec: kibanaeckv1alpha1.DataViewSpec{
					TargetConfig: kibanaeckv1alpha1.CommonKibanaConfig{
						KibanaInstance:          "my-kibana",
						KibanaInstanceNamespace: "kibana-system",
					},
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "events-*"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dataView)).Should(Succeed())

			dataViewLookupKey := types.NamespacedName{Name: dataViewName, Namespace: DataViewNamespace}
			createdDataView := &kibanaeckv1alpha1.DataView{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dataViewLookupKey, createdDataView)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdDataView.Spec.TargetConfig.KibanaInstance).Should(Equal("my-kibana"))
			Expect(createdDataView.Spec.TargetConfig.KibanaInstanceNamespace).Should(Equal("kibana-system"))
		})
	})

	Context("When updating a DataView", func() {
		It("Should update the DataView body successfully", func() {
			ctx := context.Background()

			dataViewName := "test-dataview-update"
			dataView := &kibanaeckv1alpha1.DataView{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "DataView",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataViewName,
					Namespace: DataViewNamespace,
				},
				Spec: kibanaeckv1alpha1.DataViewSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "old-pattern-*"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dataView)).Should(Succeed())

			dataViewLookupKey := types.NamespacedName{Name: dataViewName, Namespace: DataViewNamespace}
			createdDataView := &kibanaeckv1alpha1.DataView{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dataViewLookupKey, createdDataView)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdDataView.Spec.Body = `{"title": "new-pattern-*"}`
			Expect(k8sClient.Update(ctx, createdDataView)).Should(Succeed())

			updatedDataView := &kibanaeckv1alpha1.DataView{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, dataViewLookupKey, updatedDataView)
				if err != nil {
					return false
				}
				return updatedDataView.Spec.Body != dataView.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedDataView.Spec.Body).Should(ContainSubstring("new-pattern-*"))
		})
	})

	Context("When deleting a DataView", func() {
		It("Should delete the DataView resource successfully", func() {
			ctx := context.Background()

			dataViewName := "test-dataview-delete"
			dataView := &kibanaeckv1alpha1.DataView{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "DataView",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataViewName,
					Namespace: DataViewNamespace,
				},
				Spec: kibanaeckv1alpha1.DataViewSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "delete-me-*"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dataView)).Should(Succeed())

			dataViewLookupKey := types.NamespacedName{Name: dataViewName, Namespace: DataViewNamespace}
			createdDataView := &kibanaeckv1alpha1.DataView{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dataViewLookupKey, createdDataView)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdDataView)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, dataViewLookupKey, &kibanaeckv1alpha1.DataView{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
