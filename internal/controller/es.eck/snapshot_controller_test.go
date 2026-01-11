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

var _ = Describe("SnapshotRepository Controller", func() {
	const (
		SnapshotRepoNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a SnapshotRepository", func() {
		It("Should create the SnapshotRepository resource successfully", func() {
			ctx := context.Background()

			repoName := "test-snapshot-repo"
			repo := &eseckv1alpha1.SnapshotRepository{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotRepository",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      repoName,
					Namespace: SnapshotRepoNamespace,
				},
				Spec: eseckv1alpha1.SnapshotRepositorySpec{
					Body: `{
						"type": "fs",
						"settings": {
							"location": "/backup/snapshots"
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, repo)).Should(Succeed())

			repoLookupKey := types.NamespacedName{Name: repoName, Namespace: SnapshotRepoNamespace}
			createdRepo := &eseckv1alpha1.SnapshotRepository{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, repoLookupKey, createdRepo)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdRepo.Spec.Body).Should(ContainSubstring("type"))
			Expect(createdRepo.Spec.Body).Should(ContainSubstring("fs"))
		})

		It("Should create SnapshotRepository with S3 settings", func() {
			ctx := context.Background()

			repoName := "test-snapshot-repo-s3"
			repo := &eseckv1alpha1.SnapshotRepository{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotRepository",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      repoName,
					Namespace: SnapshotRepoNamespace,
				},
				Spec: eseckv1alpha1.SnapshotRepositorySpec{
					Body: `{
						"type": "s3",
						"settings": {
							"bucket": "my-bucket",
							"base_path": "snapshots",
							"compress": true
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, repo)).Should(Succeed())

			repoLookupKey := types.NamespacedName{Name: repoName, Namespace: SnapshotRepoNamespace}
			createdRepo := &eseckv1alpha1.SnapshotRepository{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, repoLookupKey, createdRepo)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdRepo.Spec.Body).Should(ContainSubstring("s3"))
			Expect(createdRepo.Spec.Body).Should(ContainSubstring("my-bucket"))
		})

		It("Should create SnapshotRepository with target instance config", func() {
			ctx := context.Background()

			repoName := "test-snapshot-repo-target"
			repo := &eseckv1alpha1.SnapshotRepository{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotRepository",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      repoName,
					Namespace: SnapshotRepoNamespace,
				},
				Spec: eseckv1alpha1.SnapshotRepositorySpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance: "my-elasticsearch",
					},
					Body: `{
						"type": "fs",
						"settings": {
							"location": "/backup"
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, repo)).Should(Succeed())

			repoLookupKey := types.NamespacedName{Name: repoName, Namespace: SnapshotRepoNamespace}
			createdRepo := &eseckv1alpha1.SnapshotRepository{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, repoLookupKey, createdRepo)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdRepo.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
		})
	})

	Context("When deleting a SnapshotRepository", func() {
		It("Should delete the SnapshotRepository resource successfully", func() {
			ctx := context.Background()

			repoName := "test-snapshot-repo-delete"
			repo := &eseckv1alpha1.SnapshotRepository{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotRepository",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      repoName,
					Namespace: SnapshotRepoNamespace,
				},
				Spec: eseckv1alpha1.SnapshotRepositorySpec{
					Body: `{"type": "fs", "settings": {"location": "/backup"}}`,
				},
			}

			Expect(k8sClient.Create(ctx, repo)).Should(Succeed())

			repoLookupKey := types.NamespacedName{Name: repoName, Namespace: SnapshotRepoNamespace}
			createdRepo := &eseckv1alpha1.SnapshotRepository{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, repoLookupKey, createdRepo)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdRepo)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, repoLookupKey, &eseckv1alpha1.SnapshotRepository{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = Describe("SnapshotLifecyclePolicy Controller", func() {
	const (
		SLPNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a SnapshotLifecyclePolicy", func() {
		It("Should create the SnapshotLifecyclePolicy resource successfully", func() {
			ctx := context.Background()

			slpName := "test-slp"
			slp := &eseckv1alpha1.SnapshotLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      slpName,
					Namespace: SLPNamespace,
				},
				Spec: eseckv1alpha1.SnapshotLifecyclePolicySpec{
					Body: `{
						"schedule": "0 30 1 * * ?",
						"name": "<daily-snap-{now/d}>",
						"repository": "my_repository",
						"config": {
							"indices": ["*"]
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, slp)).Should(Succeed())

			slpLookupKey := types.NamespacedName{Name: slpName, Namespace: SLPNamespace}
			createdSLP := &eseckv1alpha1.SnapshotLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, slpLookupKey, createdSLP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSLP.Spec.Body).Should(ContainSubstring("schedule"))
			Expect(createdSLP.Spec.Body).Should(ContainSubstring("repository"))
		})

		It("Should create SnapshotLifecyclePolicy with retention", func() {
			ctx := context.Background()

			slpName := "test-slp-retention"
			slp := &eseckv1alpha1.SnapshotLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      slpName,
					Namespace: SLPNamespace,
				},
				Spec: eseckv1alpha1.SnapshotLifecyclePolicySpec{
					Body: `{
						"schedule": "0 30 1 * * ?",
						"name": "<daily-snap-{now/d}>",
						"repository": "my_repository",
						"config": {
							"indices": ["*"],
							"include_global_state": true
						},
						"retention": {
							"expire_after": "30d",
							"min_count": 5,
							"max_count": 50
						}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, slp)).Should(Succeed())

			slpLookupKey := types.NamespacedName{Name: slpName, Namespace: SLPNamespace}
			createdSLP := &eseckv1alpha1.SnapshotLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, slpLookupKey, createdSLP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSLP.Spec.Body).Should(ContainSubstring("retention"))
			Expect(createdSLP.Spec.Body).Should(ContainSubstring("expire_after"))
		})

		It("Should create SnapshotLifecyclePolicy with target instance config", func() {
			ctx := context.Background()

			slpName := "test-slp-target"
			slp := &eseckv1alpha1.SnapshotLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      slpName,
					Namespace: SLPNamespace,
				},
				Spec: eseckv1alpha1.SnapshotLifecyclePolicySpec{
					TargetConfig: eseckv1alpha1.CommonElasticsearchConfig{
						ElasticsearchInstance: "my-elasticsearch",
					},
					Body: `{
						"schedule": "0 0 * * * ?",
						"name": "<hourly-snap-{now/H}>",
						"repository": "test_repo",
						"config": {}
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, slp)).Should(Succeed())

			slpLookupKey := types.NamespacedName{Name: slpName, Namespace: SLPNamespace}
			createdSLP := &eseckv1alpha1.SnapshotLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, slpLookupKey, createdSLP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSLP.Spec.TargetConfig.ElasticsearchInstance).Should(Equal("my-elasticsearch"))
		})
	})

	Context("When deleting a SnapshotLifecyclePolicy", func() {
		It("Should delete the SnapshotLifecyclePolicy resource successfully", func() {
			ctx := context.Background()

			slpName := "test-slp-delete"
			slp := &eseckv1alpha1.SnapshotLifecyclePolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "es.eck.github.com/v1alpha1",
					Kind:       "SnapshotLifecyclePolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      slpName,
					Namespace: SLPNamespace,
				},
				Spec: eseckv1alpha1.SnapshotLifecyclePolicySpec{
					Body: `{"schedule": "0 0 * * * ?", "name": "test", "repository": "test", "config": {}}`,
				},
			}

			Expect(k8sClient.Create(ctx, slp)).Should(Succeed())

			slpLookupKey := types.NamespacedName{Name: slpName, Namespace: SLPNamespace}
			createdSLP := &eseckv1alpha1.SnapshotLifecyclePolicy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, slpLookupKey, createdSLP)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, createdSLP)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, slpLookupKey, &eseckv1alpha1.SnapshotLifecyclePolicy{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
