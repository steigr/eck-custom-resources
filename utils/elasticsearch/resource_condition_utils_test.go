package elasticsearch

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test condition types for testing
var testConditionTypes = ResourceConditions{
	InitialDeploymentType: "InitialDeployment",
	LastUpdateType:        "LastUpdate",
	ReasonSucceeded:       "Succeeded",
	ReasonFailed:          "Failed",
	ReasonPending:         "Pending",
	ReasonBlocked:         "Blocked",
}

func TestIsInitialDeployment(t *testing.T) {
	tests := []struct {
		name       string
		conditions []metav1.Condition
		expected   bool
	}{
		{
			name:       "empty conditions - is initial deployment",
			conditions: []metav1.Condition{},
			expected:   true,
		},
		{
			name:       "nil conditions - is initial deployment",
			conditions: nil,
			expected:   true,
		},
		{
			name: "has InitialDeployment condition - not initial deployment",
			conditions: []metav1.Condition{
				{
					Type:   "InitialDeployment",
					Status: metav1.ConditionTrue,
					Reason: "Succeeded",
				},
			},
			expected: false,
		},
		{
			name: "has other conditions but not InitialDeployment - is initial deployment",
			conditions: []metav1.Condition{
				{
					Type:   "LastUpdate",
					Status: metav1.ConditionTrue,
					Reason: "Succeeded",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInitialDeployment(tt.conditions, testConditionTypes)
			if result != tt.expected {
				t.Errorf("IsInitialDeployment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCheckExternalModification(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	earlier := now.Add(-1 * time.Hour)

	tests := []struct {
		name              string
		conditions        []metav1.Condition
		esMeta            map[string]any
		expectModified    bool
		expectESUpdatedAt *time.Time
	}{
		{
			name:           "no LastUpdate condition - not modified",
			conditions:     []metav1.Condition{},
			esMeta:         map[string]any{"updated_at": now.Format(time.RFC3339)},
			expectModified: false,
		},
		{
			name: "no updated_at in meta - not modified",
			conditions: []metav1.Condition{
				{
					Type:               "LastUpdate",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
				},
			},
			esMeta:         map[string]any{},
			expectModified: false,
		},
		{
			name: "nil meta - not modified",
			conditions: []metav1.Condition{
				{
					Type:               "LastUpdate",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
				},
			},
			esMeta:         nil,
			expectModified: false,
		},
		{
			name: "timestamps match - not modified",
			conditions: []metav1.Condition{
				{
					Type:               "LastUpdate",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
				},
			},
			esMeta:         map[string]any{"updated_at": now.Format(time.RFC3339)},
			expectModified: false,
		},
		{
			name: "timestamps differ - modified externally",
			conditions: []metav1.Condition{
				{
					Type:               "LastUpdate",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(earlier),
				},
			},
			esMeta:            map[string]any{"updated_at": now.Format(time.RFC3339)},
			expectModified:    true,
			expectESUpdatedAt: &now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckExternalModification(tt.conditions, tt.esMeta, testConditionTypes)

			if result.Modified != tt.expectModified {
				t.Errorf("CheckExternalModification().Modified = %v, want %v", result.Modified, tt.expectModified)
			}

			if tt.expectModified {
				if result.ESUpdatedAt == nil {
					t.Error("CheckExternalModification().ESUpdatedAt is nil, expected non-nil")
				} else if !result.ESUpdatedAt.Equal(*tt.expectESUpdatedAt) {
					t.Errorf("CheckExternalModification().ESUpdatedAt = %v, want %v", result.ESUpdatedAt, tt.expectESUpdatedAt)
				}

				if result.ConditionToSet == nil {
					t.Error("CheckExternalModification().ConditionToSet is nil, expected non-nil")
				} else {
					if result.ConditionToSet.Type != testConditionTypes.LastUpdateType {
						t.Errorf("ConditionToSet.Type = %v, want %v", result.ConditionToSet.Type, testConditionTypes.LastUpdateType)
					}
					if result.ConditionToSet.Status != metav1.ConditionFalse {
						t.Errorf("ConditionToSet.Status = %v, want %v", result.ConditionToSet.Status, metav1.ConditionFalse)
					}
					if result.ConditionToSet.Reason != testConditionTypes.ReasonBlocked {
						t.Errorf("ConditionToSet.Reason = %v, want %v", result.ConditionToSet.Reason, testConditionTypes.ReasonBlocked)
					}
				}
			}
		})
	}
}

func TestSetSuccessConditions(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	earlier := now.Add(-1 * time.Hour)

	tests := []struct {
		name                    string
		initialConditions       []metav1.Condition
		esMeta                  map[string]any
		isInitialDeployment     bool
		expectConditionCount    int
		expectInitialDeployment bool
		expectLastUpdate        bool
	}{
		{
			name:                    "initial deployment with timestamps from ES",
			initialConditions:       []metav1.Condition{},
			esMeta:                  map[string]any{"created_at": earlier.Format(time.RFC3339), "updated_at": now.Format(time.RFC3339)},
			isInitialDeployment:     true,
			expectConditionCount:    2,
			expectInitialDeployment: true,
			expectLastUpdate:        true,
		},
		{
			name:                    "initial deployment without timestamps in ES",
			initialConditions:       []metav1.Condition{},
			esMeta:                  map[string]any{},
			isInitialDeployment:     true,
			expectConditionCount:    2,
			expectInitialDeployment: true,
			expectLastUpdate:        true,
		},
		{
			name: "update (not initial deployment)",
			initialConditions: []metav1.Condition{
				{
					Type:   "InitialDeployment",
					Status: metav1.ConditionTrue,
					Reason: "Succeeded",
				},
			},
			esMeta:                  map[string]any{"updated_at": now.Format(time.RFC3339)},
			isInitialDeployment:     false,
			expectConditionCount:    2,
			expectInitialDeployment: true,
			expectLastUpdate:        true,
		},
		{
			name:                    "nil meta",
			initialConditions:       []metav1.Condition{},
			esMeta:                  nil,
			isInitialDeployment:     true,
			expectConditionCount:    2,
			expectInitialDeployment: true,
			expectLastUpdate:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions := make([]metav1.Condition, len(tt.initialConditions))
			copy(conditions, tt.initialConditions)

			SetSuccessConditions(&conditions, tt.esMeta, tt.isInitialDeployment, testConditionTypes)

			if len(conditions) != tt.expectConditionCount {
				t.Errorf("Expected %d conditions, got %d", tt.expectConditionCount, len(conditions))
			}

			// Check InitialDeployment condition
			var foundInitialDeployment, foundLastUpdate bool
			for _, c := range conditions {
				if c.Type == testConditionTypes.InitialDeploymentType {
					foundInitialDeployment = true
					if c.Status != metav1.ConditionTrue {
						t.Errorf("InitialDeployment condition Status = %v, want %v", c.Status, metav1.ConditionTrue)
					}
					if c.Reason != testConditionTypes.ReasonSucceeded {
						t.Errorf("InitialDeployment condition Reason = %v, want %v", c.Reason, testConditionTypes.ReasonSucceeded)
					}
				}
				if c.Type == testConditionTypes.LastUpdateType {
					foundLastUpdate = true
					if c.Status != metav1.ConditionTrue {
						t.Errorf("LastUpdate condition Status = %v, want %v", c.Status, metav1.ConditionTrue)
					}
					if c.Reason != testConditionTypes.ReasonSucceeded {
						t.Errorf("LastUpdate condition Reason = %v, want %v", c.Reason, testConditionTypes.ReasonSucceeded)
					}
				}
			}

			if tt.expectInitialDeployment && !foundInitialDeployment {
				t.Error("Expected InitialDeployment condition but not found")
			}
			if tt.expectLastUpdate && !foundLastUpdate {
				t.Error("Expected LastUpdate condition but not found")
			}
		})
	}
}

func TestSetFailureConditions(t *testing.T) {
	tests := []struct {
		name                    string
		initialConditions       []metav1.Condition
		isInitialDeployment     bool
		errMsg                  string
		expectConditionCount    int
		expectInitialDeployment bool
		expectLastUpdate        bool
	}{
		{
			name:                    "initial deployment failure",
			initialConditions:       []metav1.Condition{},
			isInitialDeployment:     true,
			errMsg:                  "connection refused",
			expectConditionCount:    2,
			expectInitialDeployment: true,
			expectLastUpdate:        true,
		},
		{
			name: "update failure (not initial deployment)",
			initialConditions: []metav1.Condition{
				{
					Type:   "InitialDeployment",
					Status: metav1.ConditionTrue,
					Reason: "Succeeded",
				},
			},
			isInitialDeployment:     false,
			errMsg:                  "timeout",
			expectConditionCount:    2,
			expectInitialDeployment: true,
			expectLastUpdate:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions := make([]metav1.Condition, len(tt.initialConditions))
			copy(conditions, tt.initialConditions)

			SetFailureConditions(&conditions, tt.isInitialDeployment, testConditionTypes, tt.errMsg)

			if len(conditions) != tt.expectConditionCount {
				t.Errorf("Expected %d conditions, got %d", tt.expectConditionCount, len(conditions))
			}

			// Check conditions
			for _, c := range conditions {
				if c.Type == testConditionTypes.InitialDeploymentType && tt.isInitialDeployment {
					if c.Status != metav1.ConditionFalse {
						t.Errorf("InitialDeployment condition Status = %v, want %v", c.Status, metav1.ConditionFalse)
					}
					if c.Reason != testConditionTypes.ReasonFailed {
						t.Errorf("InitialDeployment condition Reason = %v, want %v", c.Reason, testConditionTypes.ReasonFailed)
					}
					if c.Message != "Initial deployment failed: "+tt.errMsg {
						t.Errorf("InitialDeployment condition Message = %v, want %v", c.Message, "Initial deployment failed: "+tt.errMsg)
					}
				}
				if c.Type == testConditionTypes.LastUpdateType {
					if c.Status != metav1.ConditionFalse {
						t.Errorf("LastUpdate condition Status = %v, want %v", c.Status, metav1.ConditionFalse)
					}
					if c.Reason != testConditionTypes.ReasonFailed {
						t.Errorf("LastUpdate condition Reason = %v, want %v", c.Reason, testConditionTypes.ReasonFailed)
					}
					if c.Message != "Update failed: "+tt.errMsg {
						t.Errorf("LastUpdate condition Message = %v, want %v", c.Message, "Update failed: "+tt.errMsg)
					}
				}
			}
		})
	}
}

func TestSetSuccessConditions_UsesESTimestamps(t *testing.T) {
	createdAt := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 10, 14, 0, 0, 0, time.UTC)

	esMeta := map[string]any{
		"created_at": createdAt.Format(time.RFC3339),
		"updated_at": updatedAt.Format(time.RFC3339),
	}

	conditions := []metav1.Condition{}
	SetSuccessConditions(&conditions, esMeta, true, testConditionTypes)

	for _, c := range conditions {
		if c.Type == testConditionTypes.InitialDeploymentType {
			if !c.LastTransitionTime.Time.Equal(createdAt) {
				t.Errorf("InitialDeployment LastTransitionTime = %v, want %v", c.LastTransitionTime.Time, createdAt)
			}
		}
		if c.Type == testConditionTypes.LastUpdateType {
			if !c.LastTransitionTime.Time.Equal(updatedAt) {
				t.Errorf("LastUpdate LastTransitionTime = %v, want %v", c.LastTransitionTime.Time, updatedAt)
			}
		}
	}
}
