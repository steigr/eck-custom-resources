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
	"eck-custom-resources/utils/template"
	"fmt"
	"time"

	configv2 "eck-custom-resources/api/config/v2"
	"eck-custom-resources/utils"
	esutils "eck-custom-resources/utils/elasticsearch"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	eseckv1alpha1 "eck-custom-resources/api/es.eck/v1alpha1"
)

// IngestPipelineReconciler reconciles a IngestPipeline object
type IngestPipelineReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ProjectConfig configv2.ProjectConfigSpec
	Recorder      record.EventRecorder
	RestConfig    *rest.Config
}

//+kubebuilder:rbac:groups=es.eck.github.com,resources=ingestpipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=es.eck.github.com,resources=ingestpipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=es.eck.github.com,resources=ingestpipelines/finalizers,verbs=update

func (r *IngestPipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	finalizer := "ingestpipelines.es.eck.github.com/finalizer"

	var ingestPipeline eseckv1alpha1.IngestPipeline
	if err := r.Get(ctx, req.NamespacedName, &ingestPipeline); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	targetInstance, err := esutils.GetElasticsearchTargetInstance(r.Client, ctx, r.Recorder, &ingestPipeline, r.ProjectConfig.Elasticsearch, ingestPipeline.Spec.TargetConfig, req.Namespace)
	if err != nil {
		return utils.GetRequeueResult(), err
	}

	if !targetInstance.Enabled {
		logger.Info("Elasticsearch reconciler disabled, not reconciling.", "Resource", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	targetInstanceNamespace := req.Namespace
	if ingestPipeline.Spec.TargetConfig.ElasticsearchInstanceNamespace != "" {
		targetInstanceNamespace = ingestPipeline.Spec.TargetConfig.ElasticsearchInstanceNamespace
	}

	esClient, createClientErr := esutils.GetElasticsearchClient(r.Client, ctx, *targetInstance, req, targetInstanceNamespace)
	if createClientErr != nil {
		logger.Error(createClientErr, "Failed to create Elasticsearch client")
		return utils.GetRequeueResult(), client.IgnoreNotFound(createClientErr)
	}

	// Handle deletion
	if !ingestPipeline.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&ingestPipeline, finalizer) {
			logger.Info("Deleting object", "ingestPipeline", ingestPipeline.Name)
			if _, err := esutils.DeleteIngestPipeline(esClient, req.Name); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&ingestPipeline, finalizer)
			if err := r.Update(ctx, &ingestPipeline); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Handle create/update
	logger.Info("Creating/Updating object", "ingestPipeline", ingestPipeline.Name)

	// Determine the body to use - either rendered from template or original
	body, err := template.FetchAndRenderTemplate(
		r.Client,
		ctx,
		ingestPipeline.Spec.Template,
		ingestPipeline.Spec.Body,
		req.Namespace,
		r.RestConfig,
	)
	if err != nil {
		r.Recorder.Event(&ingestPipeline, "Warning", "TemplateRenderError",
			fmt.Sprintf("Failed to render template: %s", err.Error()))
		return utils.GetRequeueResult(), err
	}

	// Create or update the Ingest pipeline in Elasticsearch
	logger.Info("Creating/Updating Ingest pipeline", "id", req.Name)

	// Define condition types for this resource
	conditionTypes := esutils.ResourceConditions{
		InitialDeploymentType: eseckv1alpha1.IngestPipelineConditionTypeInitialDeployment,
		LastUpdateType:        eseckv1alpha1.IngestPipelineConditionTypeLastUpdate,
		ReasonSucceeded:       eseckv1alpha1.IngestPipelineReasonSucceeded,
		ReasonFailed:          eseckv1alpha1.IngestPipelineReasonFailed,
		ReasonPending:         eseckv1alpha1.IngestPipelineReasonPending,
		ReasonBlocked:         eseckv1alpha1.IngestPipelineReasonBlocked,
	}

	// Check if this is the initial deployment
	isInitialDeployment := esutils.IsInitialDeployment(ingestPipeline.Status.Conditions, conditionTypes)

	// If not initial deployment and UpdateMode is not Overwrite, check if the pipeline was modified externally in Elasticsearch
	if !isInitialDeployment && ingestPipeline.Spec.UpdatePolicy.UpdateMode != eseckv1alpha1.UpdateModeOverwrite {
		pipeline, err := esutils.GetIngestPipeline(esClient, ingestPipeline.Name)
		if err != nil {
			logger.Error(err, "Failed to get ingest pipeline from Elasticsearch")
			// Continue with update if we can't check the timestamp
		} else if pipeline != nil {
			modResult := esutils.CheckExternalModification(ingestPipeline.Status.Conditions, pipeline.Meta, conditionTypes)
			if modResult.Modified {
				logger.Info("Ingest pipeline was modified externally in Elasticsearch, skipping update",
					"esUpdatedAt", modResult.ESUpdatedAt.Format(time.RFC3339))
				r.Recorder.Event(&ingestPipeline, "Warning", "ExternalModification",
					fmt.Sprintf("Ingest pipeline %s was modified externally in Elasticsearch, skipping update", ingestPipeline.Name))

				meta.SetStatusCondition(&ingestPipeline.Status.Conditions, *modResult.ConditionToSet)
				ingestPipeline.Status.ObservedGeneration = ingestPipeline.Generation
				if statusErr := r.Status().Update(ctx, &ingestPipeline); statusErr != nil {
					logger.Error(statusErr, "Failed to update IngestPipeline status")
				}
				return ctrl.Result{}, nil
			}
		}
	}

	result, err := esutils.UpsertIngestPipeline(esClient, ingestPipeline, body)

	if err == nil {
		r.Recorder.Event(&ingestPipeline, "Normal", "Created",
			fmt.Sprintf("Created/Updated %s/%s %s", ingestPipeline.APIVersion, ingestPipeline.Kind, ingestPipeline.Name))

		// Get the pipeline to extract timestamps and set success conditions
		pipeline, pipelineErr := esutils.GetIngestPipeline(esClient, ingestPipeline.Name)
		var esMeta map[string]any
		if pipelineErr == nil && pipeline != nil {
			esMeta = pipeline.Meta
		}
		esutils.SetSuccessConditions(&ingestPipeline.Status.Conditions, esMeta, isInitialDeployment, conditionTypes)
	} else {
		r.Recorder.Event(&ingestPipeline, "Warning", "Failed to create/update",
			fmt.Sprintf("Failed to create/update %s/%s %s: %s", ingestPipeline.APIVersion, ingestPipeline.Kind, ingestPipeline.Name, err.Error()))

		esutils.SetFailureConditions(&ingestPipeline.Status.Conditions, isInitialDeployment, conditionTypes, err.Error())
	}

	// Update status with observed generation
	ingestPipeline.Status.ObservedGeneration = ingestPipeline.Generation
	if statusErr := r.Status().Update(ctx, &ingestPipeline); statusErr != nil {
		logger.Error(statusErr, "Failed to update IngestPipeline status")
		// Don't return error here, continue with the main operation result
	}

	if err := r.addFinalizer(&ingestPipeline, finalizer, ctx); err != nil {
		return ctrl.Result{}, err
	}
	return result, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngestPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eseckv1alpha1.IngestPipeline{}).
		WithEventFilter(utils.CommonEventFilter()).
		Complete(r)
}

func (r *IngestPipelineReconciler) addFinalizer(o client.Object, finalizer string, ctx context.Context) error {
	if !controllerutil.ContainsFinalizer(o, finalizer) {
		controllerutil.AddFinalizer(o, finalizer)
		if err := r.Update(ctx, o); err != nil {
			return err
		}
	}
	return nil
}
