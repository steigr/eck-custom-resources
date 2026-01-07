package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	configv2 "eck-custom-resources/api/config/v2"

	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// MockObject implements runtime.Object for testing
type MockObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (m *MockObject) DeepCopyObject() runtime.Object {
	return &MockObject{
		TypeMeta:   m.TypeMeta,
		ObjectMeta: *m.ObjectMeta.DeepCopy(),
	}
}

func TestGetRequeueResult(t *testing.T) {
	result := GetRequeueResult()

	if !result.Requeue {
		t.Errorf("GetRequeueResult().Requeue = %v, want true", result.Requeue)
	}

	expectedDuration := time.Duration(time.Minute)
	if result.RequeueAfter != expectedDuration {
		t.Errorf("GetRequeueResult().RequeueAfter = %v, want %v", result.RequeueAfter, expectedDuration)
	}
}

func TestRecordError(t *testing.T) {
	recorder := record.NewFakeRecorder(10)
	obj := &MockObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-object",
			Namespace: "default",
		},
	}

	errorEvent := ErrorEvent{
		Event: Event{
			Object:  obj,
			Name:    "test-resource",
			Reason:  "TestError",
			Message: "Test error message",
		},
		Err: errors.New("something went wrong"),
	}

	RecordError(recorder, errorEvent)

	select {
	case event := <-recorder.Events:
		expectedContains := "Warning"
		if event == "" {
			t.Error("RecordError() did not record any event")
		}
		if len(event) == 0 {
			t.Errorf("RecordError() recorded empty event")
		}
		// Check that it's a warning event
		if !containsString(event, expectedContains) {
			t.Errorf("RecordError() event = %v, should contain %v", event, expectedContains)
		}
	default:
		t.Error("RecordError() did not record any event")
	}
}

func TestRecordSuccess(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantMsg string
	}{
		{
			name: "with custom message",
			event: Event{
				Object: &MockObject{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-object",
						Namespace: "default",
					},
				},
				Name:    "test-resource",
				Reason:  "Created",
				Message: "Custom message",
			},
			wantMsg: "Custom message for test-resource",
		},
		{
			name: "without custom message",
			event: Event{
				Object: &MockObject{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-object",
						Namespace: "default",
					},
				},
				Name:   "test-resource",
				Reason: "Created",
			},
			wantMsg: "Created successful for test-resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := record.NewFakeRecorder(10)
			RecordSuccess(recorder, tt.event)

			select {
			case event := <-recorder.Events:
				if !containsString(event, "Normal") {
					t.Errorf("RecordSuccess() event = %v, should contain 'Normal'", event)
				}
			default:
				t.Error("RecordSuccess() did not record any event")
			}
		})
	}
}

func TestRecordEventAndReturn(t *testing.T) {
	tests := []struct {
		name      string
		res       ctrl.Result
		err       error
		event     Event
		wantRes   ctrl.Result
		wantErr   error
		wantEvent string
	}{
		{
			name: "success case",
			res:  ctrl.Result{},
			err:  nil,
			event: Event{
				Object: &MockObject{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-object",
						Namespace: "default",
					},
				},
				Name:   "test-resource",
				Reason: "Created",
			},
			wantRes:   ctrl.Result{},
			wantErr:   nil,
			wantEvent: "Normal",
		},
		{
			name: "error case",
			res:  ctrl.Result{Requeue: true},
			err:  errors.New("test error"),
			event: Event{
				Object: &MockObject{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-object",
						Namespace: "default",
					},
				},
				Name:    "test-resource",
				Reason:  "Failed",
				Message: "Operation failed",
			},
			wantRes:   ctrl.Result{Requeue: true},
			wantErr:   errors.New("test error"),
			wantEvent: "Warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := record.NewFakeRecorder(10)
			gotRes, gotErr := RecordEventAndReturn(tt.res, tt.err, recorder, tt.event)

			if gotRes != tt.wantRes {
				t.Errorf("RecordEventAndReturn() res = %v, want %v", gotRes, tt.wantRes)
			}

			if (gotErr == nil) != (tt.wantErr == nil) {
				t.Errorf("RecordEventAndReturn() err = %v, want %v", gotErr, tt.wantErr)
			}

			select {
			case event := <-recorder.Events:
				if !containsString(event, tt.wantEvent) {
					t.Errorf("RecordEventAndReturn() event = %v, should contain %v", event, tt.wantEvent)
				}
			default:
				t.Error("RecordEventAndReturn() did not record any event")
			}
		})
	}
}

func TestGetUserSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = k8sv1.AddToScheme(scheme)

	secret := &k8sv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("testuser"),
			"password": []byte("testpass"),
		},
	}

	tests := []struct {
		name       string
		namespace  string
		auth       *configv2.UsernamePasswordAuthentication
		wantErr    bool
		wantSecret bool
	}{
		{
			name:      "secret exists",
			namespace: "default",
			auth: &configv2.UsernamePasswordAuthentication{
				SecretName: "test-secret",
			},
			wantErr:    false,
			wantSecret: true,
		},
		{
			name:      "secret does not exist",
			namespace: "default",
			auth: &configv2.UsernamePasswordAuthentication{
				SecretName: "nonexistent-secret",
			},
			wantErr:    true,
			wantSecret: false,
		},
		{
			name:      "wrong namespace",
			namespace: "other-namespace",
			auth: &configv2.UsernamePasswordAuthentication{
				SecretName: "test-secret",
			},
			wantErr:    true,
			wantSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(secret).
				Build()

			var gotSecret k8sv1.Secret
			err := GetUserSecret(fakeClient, context.Background(), tt.namespace, tt.auth, &gotSecret)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantSecret && gotSecret.Name != "test-secret" {
				t.Errorf("GetUserSecret() secret name = %v, want test-secret", gotSecret.Name)
			}
		})
	}
}

func TestGetCertificateSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = k8sv1.AddToScheme(scheme)

	secret := &k8sv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cert-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"ca.crt": []byte("-----BEGIN CERTIFICATE-----"),
		},
	}

	tests := []struct {
		name        string
		namespace   string
		certificate *configv2.PublicCertificate
		wantErr     bool
		wantSecret  bool
	}{
		{
			name:      "certificate secret exists",
			namespace: "default",
			certificate: &configv2.PublicCertificate{
				SecretName: "cert-secret",
			},
			wantErr:    false,
			wantSecret: true,
		},
		{
			name:      "certificate secret does not exist",
			namespace: "default",
			certificate: &configv2.PublicCertificate{
				SecretName: "nonexistent-cert",
			},
			wantErr:    true,
			wantSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(secret).
				Build()

			var gotSecret k8sv1.Secret
			err := GetCertificateSecret(fakeClient, context.Background(), tt.namespace, tt.certificate, &gotSecret)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCertificateSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantSecret && gotSecret.Name != "cert-secret" {
				t.Errorf("GetCertificateSecret() secret name = %v, want cert-secret", gotSecret.Name)
			}
		})
	}
}

func TestCommonEventFilter(t *testing.T) {
	filter := CommonEventFilter()

	tests := []struct {
		name          string
		oldGeneration int64
		newGeneration int64
		want          bool
	}{
		{
			name:          "generation changed - should process",
			oldGeneration: 1,
			newGeneration: 2,
			want:          true,
		},
		{
			name:          "generation unchanged - should skip",
			oldGeneration: 1,
			newGeneration: 1,
			want:          false,
		},
		{
			name:          "new object - generation 0 to 1",
			oldGeneration: 0,
			newGeneration: 1,
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldObj := &MockObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-object",
					Generation: tt.oldGeneration,
				},
			}
			newObj := &MockObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-object",
					Generation: tt.newGeneration,
				},
			}

			updateEvent := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}

			got := filter.UpdateFunc(updateEvent)
			if got != tt.want {
				t.Errorf("CommonEventFilter().UpdateFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
