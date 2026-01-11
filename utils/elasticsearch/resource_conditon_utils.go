package elasticsearch

import (
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceConditions holds the condition type names for a resource
type ResourceConditions struct {
	InitialDeploymentType string
	LastUpdateType        string
	ReasonSucceeded       string
	ReasonFailed          string
	ReasonPending         string
	ReasonBlocked         string
}

// ExternalModificationResult contains the result of checking for external modifications
type ExternalModificationResult struct {
	Modified       bool
	ESUpdatedAt    *time.Time
	ConditionToSet *metav1.Condition
}

// CheckExternalModification checks if a resource was modified externally in Elasticsearch
// by comparing the updated_at timestamp from ES with the LastTransitionTime of the LastUpdate condition.
// Returns ExternalModificationResult indicating if modification was detected.
func CheckExternalModification(
	conditions []metav1.Condition,
	esMeta map[string]any,
	conditionTypes ResourceConditions,
) ExternalModificationResult {
	result := ExternalModificationResult{Modified: false}

	lastUpdateCondition := meta.FindStatusCondition(conditions, conditionTypes.LastUpdateType)
	if lastUpdateCondition == nil {
		return result
	}

	esUpdatedAt, _ := GetResourceUpdatedAt(esMeta)
	if esUpdatedAt == nil {
		return result
	}

	result.ESUpdatedAt = esUpdatedAt

	// Compare the timestamps - if ES timestamp doesn't match our last update, it was modified externally
	if !esUpdatedAt.Equal(lastUpdateCondition.LastTransitionTime.Time) {
		result.Modified = true
		result.ConditionToSet = &metav1.Condition{
			Type:               conditionTypes.LastUpdateType,
			Status:             metav1.ConditionFalse,
			Reason:             conditionTypes.ReasonBlocked,
			Message:            "Update blocked due to external modification in Elasticsearch (ES updated_at: " + esUpdatedAt.Format(time.RFC3339) + ")",
			LastTransitionTime: metav1.NewTime(*esUpdatedAt),
		}
	}

	return result
}

// SetSuccessConditions sets the InitialDeployment and LastUpdate conditions after a successful operation.
// Uses timestamps from Elasticsearch _meta if available, otherwise falls back to current time.
func SetSuccessConditions(
	conditions *[]metav1.Condition,
	esMeta map[string]any,
	isInitialDeployment bool,
	conditionTypes ResourceConditions,
) {
	esCreatedAt, _ := GetResourceCreatedAt(esMeta)
	esUpdatedAt, _ := GetResourceUpdatedAt(esMeta)

	if isInitialDeployment {
		var initialDeploymentTime metav1.Time
		if esCreatedAt != nil {
			initialDeploymentTime = metav1.NewTime(*esCreatedAt)
		} else {
			initialDeploymentTime = metav1.Now()
		}
		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               conditionTypes.InitialDeploymentType,
			Status:             metav1.ConditionTrue,
			Reason:             conditionTypes.ReasonSucceeded,
			Message:            "Initial deployment completed successfully",
			LastTransitionTime: initialDeploymentTime,
		})
	}

	var lastUpdateTime metav1.Time
	if esUpdatedAt != nil {
		lastUpdateTime = metav1.NewTime(*esUpdatedAt)
	} else {
		lastUpdateTime = metav1.Now()
	}
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionTypes.LastUpdateType,
		Status:             metav1.ConditionTrue,
		Reason:             conditionTypes.ReasonSucceeded,
		Message:            "Update applied successfully",
		LastTransitionTime: lastUpdateTime,
	})
}

// SetFailureConditions sets the InitialDeployment and LastUpdate conditions after a failed operation.
func SetFailureConditions(
	conditions *[]metav1.Condition,
	isInitialDeployment bool,
	conditionTypes ResourceConditions,
	errMsg string,
) {
	if isInitialDeployment {
		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               conditionTypes.InitialDeploymentType,
			Status:             metav1.ConditionFalse,
			Reason:             conditionTypes.ReasonFailed,
			Message:            "Initial deployment failed: " + errMsg,
			LastTransitionTime: metav1.Now(),
		})
	}
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionTypes.LastUpdateType,
		Status:             metav1.ConditionFalse,
		Reason:             conditionTypes.ReasonFailed,
		Message:            "Update failed: " + errMsg,
		LastTransitionTime: metav1.Now(),
	})
}

// IsInitialDeployment checks if this is the initial deployment by looking for the InitialDeployment condition.
func IsInitialDeployment(conditions []metav1.Condition, conditionTypes ResourceConditions) bool {
	return meta.FindStatusCondition(conditions, conditionTypes.InitialDeploymentType) == nil
}
