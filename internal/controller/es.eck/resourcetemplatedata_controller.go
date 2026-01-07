/*
Copyright 2025.

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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv2 "eck-custom-resources/api/config/v2"
	"eck-custom-resources/utils"
	esutils "eck-custom-resources/utils/elasticsearch"

	eseckv1alpha1 "eck-custom-resources/api/es.eck/v1alpha1"
)

// ResourceTemplateDataReconciler reconciles a ResourceTemplateData object
type ResourceTemplateDataReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ProjectConfig configv2.ProjectConfigSpec
	Recorder      record.EventRecorder
}

// +kubebuilder:rbac:groups=es.eck.github.com,resources=resourcetemplatedata,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=es.eck.github.com,resources=resourcetemplatedata/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=es.eck.github.com,resources=resourcetemplatedata/finalizers,verbs=update

// The reconciler must trigger all resources referencing
func (r *ResourceTemplateDataReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	finalizer := "resourcetemplatedatas.es.eck.github.com/finalizer"

	var resourceTemplateData eseckv1alpha1.ResourceTemplateData
	if err := r.Get(ctx, req.NamespacedName, &resourceTemplateData); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	targetInstance, err := esutils.GetElasticsearchTargetInstance(r.Client, ctx, r.Recorder, &resourceTemplateData, r.ProjectConfig.Elasticsearch, resourceTemplateData.Spec.TargetConfig, req.Namespace)
	if err != nil {
		return utils.GetRequeueResult(), err
	}

	if !targetInstance.Enabled {
		logger.Info("Elasticsearch reconciler disabled, not reconciling.", "Resource", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	targetInstanceNamespace := req.Namespace
	if resourceTemplateData.Spec.TargetConfig.ElasticsearchInstanceNamespace != "" {
		targetInstanceNamespace = resourceTemplateData.Spec.TargetConfig.ElasticsearchInstanceNamespace
	}
	logger.Info("Using target instance namespace", "namespace", targetInstanceNamespace)

	if resourceTemplateData.ObjectMeta.DeletionTimestamp.IsZero() {
		if err := r.triggerDependentResourcesReconcile(ctx, &resourceTemplateData, targetInstance, targetInstanceNamespace); err != nil {
			return ctrl.Result{}, err
		}
		r.Recorder.Event(&resourceTemplateData, "Normal", "Created",
			fmt.Sprintf("Created/Updated %s/%s %s", resourceTemplateData.APIVersion, resourceTemplateData.Kind, resourceTemplateData.Name))

		if err := r.addFinalizer(&resourceTemplateData, finalizer, ctx); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&resourceTemplateData, finalizer) {
			logger.Info("Deleting object", "resourceTemplateData", resourceTemplateData.Name)
			if err := r.triggerDependentResourcesReconcile(ctx, &resourceTemplateData, targetInstance, targetInstanceNamespace); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&resourceTemplateData, finalizer)
			if err := r.Update(ctx, &resourceTemplateData); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ResourceTemplateDataReconciler) addFinalizer(o client.Object, finalizer string, ctx context.Context) error {
	if !controllerutil.ContainsFinalizer(o, finalizer) {
		controllerutil.AddFinalizer(o, finalizer)
		if err := r.Update(ctx, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceTemplateDataReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eseckv1alpha1.ResourceTemplateData{}).
		WithEventFilter(utils.CommonEventFilter()).
		Complete(r)
}

// Search for all custom resources which
// reference the ResourceTemplateData and trigger reconcile
func (r *ResourceTemplateDataReconciler) triggerDependentResourcesReconcile(ctx context.Context, resourceTemplateData *eseckv1alpha1.ResourceTemplateData, targetInstance *configv2.ElasticsearchSpec, targetInstanceNamespace string) error {
	// iterate over all registered custom resources in group es.eck.github.com having .spec.template
	for _, gvk := range utils.GetRegisteredGVKsInGroupWithTemplatingSpec(r.Scheme, "es.eck.github.com") {
		dependentResources, err := utils.ListResourcesReferencingResourceTemplateData(r.Client, ctx, gvk, resourceTemplateData.Name, targetInstanceNamespace)
		if err != nil {
			return err
		}

		for _, dependentResource := range dependentResources {
			logger := log.FromContext(ctx)
			logger.Info("Triggering reconcile for dependent resource", "GVK", gvk, "Name", dependentResource.GetName(), "Namespace", dependentResource.GetNamespace())
			err = r.Client.Update(ctx, &dependentResource)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
