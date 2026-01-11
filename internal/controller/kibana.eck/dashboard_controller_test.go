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

var _ = Describe("Dashboard Controller", func() {
	const (
		DashboardNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a Dashboard", func() {
		It("Should create the Dashboard resource successfully", func() {
			ctx := context.Background()

			dashboardName := "test-dashboard"
			dashboard := &kibanaeckv1alpha1.Dashboard{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Dashboard",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dashboardName,
					Namespace: DashboardNamespace,
				},
				Spec: kibanaeckv1alpha1.DashboardSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{
							"title": "Test Dashboard",
							"panelsJSON": "[]",
							"optionsJSON": "{}"
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			dashboardLookupKey := types.NamespacedName{Name: dashboardName, Namespace: DashboardNamespace}
			createdDashboard := &kibanaeckv1alpha1.Dashboard{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdDashboard.Spec.Body).Should(ContainSubstring("Test Dashboard"))
		})

		It("Should create Dashboard with space", func() {
			ctx := context.Background()

			dashboardName := "test-dashboard-space"
			space := "my-space"
			dashboard := &kibanaeckv1alpha1.Dashboard{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Dashboard",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dashboardName,
					Namespace: DashboardNamespace,
				},
				Spec: kibanaeckv1alpha1.DashboardSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Space: &space,
						Body: `{
							"title": "Dashboard in Space",
							"panelsJSON": "[]"
						}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			dashboardLookupKey := types.NamespacedName{Name: dashboardName, Namespace: DashboardNamespace}
			createdDashboard := &kibanaeckv1alpha1.Dashboard{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdDashboard.Spec.Space).ShouldNot(BeNil())
			Expect(*createdDashboard.Spec.Space).Should(Equal("my-space"))
		})

		It("Should create Dashboard with target instance config", func() {
			ctx := context.Background()

			dashboardName := "test-dashboard-target"
			dashboard := &kibanaeckv1alpha1.Dashboard{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Dashboard",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dashboardName,
					Namespace: DashboardNamespace,
				},
				Spec: kibanaeckv1alpha1.DashboardSpec{
					TargetConfig: kibanaeckv1alpha1.CommonKibanaConfig{
						KibanaInstance: "my-kibana",
					},
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Dashboard with Target"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			dashboardLookupKey := types.NamespacedName{Name: dashboardName, Namespace: DashboardNamespace}
			createdDashboard := &kibanaeckv1alpha1.Dashboard{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdDashboard.Spec.TargetConfig.KibanaInstance).Should(Equal("my-kibana"))
		})

		It("Should create Dashboard with dependencies", func() {
			ctx := context.Background()

			dashboardName := "test-dashboard-deps"
			dashboard := &kibanaeckv1alpha1.Dashboard{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Dashboard",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dashboardName,
					Namespace: DashboardNamespace,
				},
				Spec: kibanaeckv1alpha1.DashboardSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Dashboard with Dependencies"}`,
						Dependencies: []kibanaeckv1alpha1.Dependency{
							{
								ObjectType: "visualization",
								Name:       "my-visualization",
							},
							{
								ObjectType: "index-pattern",
								Name:       "logs-*",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			dashboardLookupKey := types.NamespacedName{Name: dashboardName, Namespace: DashboardNamespace}
			createdDashboard := &kibanaeckv1alpha1.Dashboard{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(len(createdDashboard.Spec.Dependencies)).Should(Equal(2))
		})
	})

	Context("When updating a Dashboard", func() {
		It("Should update the Dashboard body successfully", func() {
			ctx := context.Background()

			dashboardName := "test-dashboard-update"
			dashboard := &kibanaeckv1alpha1.Dashboard{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Dashboard",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dashboardName,
					Namespace: DashboardNamespace,
				},
				Spec: kibanaeckv1alpha1.DashboardSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Original Title"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			dashboardLookupKey := types.NamespacedName{Name: dashboardName, Namespace: DashboardNamespace}
			createdDashboard := &kibanaeckv1alpha1.Dashboard{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the body
			createdDashboard.Spec.Body = `{"title": "Updated Title"}`
			Expect(k8sClient.Update(ctx, createdDashboard)).Should(Succeed())

			updatedDashboard := &kibanaeckv1alpha1.Dashboard{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, updatedDashboard)
				if err != nil {
					return false
				}
				return updatedDashboard.Spec.Body != dashboard.Spec.Body
			}, timeout, interval).Should(BeTrue())

			Expect(updatedDashboard.Spec.Body).Should(ContainSubstring("Updated Title"))
		})
	})

	Context("When deleting a Dashboard", func() {
		It("Should delete the Dashboard resource successfully", func() {
			ctx := context.Background()

			dashboardName := "test-dashboard-delete"
			dashboard := &kibanaeckv1alpha1.Dashboard{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kibana.eck.github.com/v1alpha1",
					Kind:       "Dashboard",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      dashboardName,
					Namespace: DashboardNamespace,
				},
				Spec: kibanaeckv1alpha1.DashboardSpec{
					SavedObject: kibanaeckv1alpha1.SavedObject{
						Body: `{"title": "Dashboard to Delete"}`,
					},
				},
			}

			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			dashboardLookupKey := types.NamespacedName{Name: dashboardName, Namespace: DashboardNamespace}
			createdDashboard := &kibanaeckv1alpha1.Dashboard{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdDashboard)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, dashboardLookupKey, &kibanaeckv1alpha1.Dashboard{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
