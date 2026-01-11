package kibana

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configv2 "eck-custom-resources/api/config/v2"
	kibanaeckv1alpha1 "eck-custom-resources/api/kibana.eck/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestFormatSavedObjectUrl(t *testing.T) {
	tests := []struct {
		name            string
		savedObjectType string
		objectName      string
		space           *string
		expected        string
	}{
		{
			name:            "nil space - dashboard",
			savedObjectType: "dashboard",
			objectName:      "my-dashboard",
			space:           nil,
			expected:        "/api/saved_objects/dashboard/my-dashboard",
		},
		{
			name:            "nil space - visualization",
			savedObjectType: "visualization",
			objectName:      "my-viz",
			space:           nil,
			expected:        "/api/saved_objects/visualization/my-viz",
		},
		{
			name:            "with space - dashboard",
			savedObjectType: "dashboard",
			objectName:      "my-dashboard",
			space:           strPtr("my-space"),
			expected:        "/s/my-space/api/saved_objects/dashboard/my-dashboard",
		},
		{
			name:            "with space - index-pattern",
			savedObjectType: "index-pattern",
			objectName:      "logs-*",
			space:           strPtr("analytics"),
			expected:        "/s/analytics/api/saved_objects/index-pattern/logs-*",
		},
		{
			name:            "default space",
			savedObjectType: "lens",
			objectName:      "my-lens",
			space:           strPtr("default"),
			expected:        "/s/default/api/saved_objects/lens/my-lens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSavedObjectUrl(tt.savedObjectType, tt.objectName, tt.space)
			if result != tt.expected {
				t.Errorf("formatSavedObjectUrl() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSavedObjectExists(t *testing.T) {
	tests := []struct {
		name             string
		savedObjectType  string
		objectName       string
		space            *string
		serverStatusCode int
		serverResponse   string
		wantExists       bool
		wantErr          bool
	}{
		{
			name:             "object exists - no space",
			savedObjectType:  "dashboard",
			objectName:       "my-dashboard",
			space:            nil,
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"id": "my-dashboard", "type": "dashboard"}`,
			wantExists:       true,
			wantErr:          false,
		},
		{
			name:             "object exists - with space",
			savedObjectType:  "visualization",
			objectName:       "my-viz",
			space:            strPtr("my-space"),
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"id": "my-viz", "type": "visualization"}`,
			wantExists:       true,
			wantErr:          false,
		},
		{
			name:             "object not found",
			savedObjectType:  "dashboard",
			objectName:       "nonexistent",
			space:            nil,
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"statusCode": 404, "error": "Not Found"}`,
			wantExists:       false,
			wantErr:          false,
		},
		{
			name:             "server error",
			savedObjectType:  "dashboard",
			objectName:       "my-dashboard",
			space:            nil,
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": "Internal Server Error"}`,
			wantExists:       false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := formatSavedObjectUrl(tt.savedObjectType, tt.objectName, tt.space)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			kClient := createTestKibanaClient(server.URL)

			exists, err := SavedObjectExists(kClient, tt.savedObjectType, tt.objectName, tt.space)

			if (err != nil) != tt.wantErr {
				t.Errorf("SavedObjectExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("SavedObjectExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestDeleteSavedObject(t *testing.T) {
	tests := []struct {
		name             string
		savedObjectType  string
		objectName       string
		space            *string
		serverStatusCode int
		wantErr          bool
	}{
		{
			name:             "successful deletion - no space",
			savedObjectType:  "dashboard",
			objectName:       "my-dashboard",
			space:            nil,
			serverStatusCode: http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "successful deletion - with space",
			savedObjectType:  "visualization",
			objectName:       "my-viz",
			space:            strPtr("my-space"),
			serverStatusCode: http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "object not found",
			savedObjectType:  "dashboard",
			objectName:       "nonexistent",
			space:            nil,
			serverStatusCode: http.StatusNotFound,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := formatSavedObjectUrl(tt.savedObjectType, tt.objectName, tt.space)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE method, got %s", r.Method)
				}

				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(`{}`))
			}))
			defer server.Close()

			kClient := createTestKibanaClient(server.URL)

			savedObject := kibanaeckv1alpha1.SavedObject{
				Space: tt.space,
				Body:  `{}`,
			}

			result, err := DeleteSavedObject(kClient, tt.savedObjectType, metav1.ObjectMeta{Name: tt.objectName}, savedObject)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSavedObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != (ctrl.Result{}) {
				t.Errorf("DeleteSavedObject() result = %v, want empty Result", result)
			}
		})
	}
}

func TestUpsertSavedObject_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call - check existence (GET)
			if r.Method != http.MethodGet {
				t.Errorf("First call: expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"statusCode": 404}`))
			return
		}

		if callCount == 2 {
			// Second call - create (POST)
			if r.Method != http.MethodPost {
				t.Errorf("Second call: expected POST method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "my-dashboard"}`))
			return
		}
	}))
	defer server.Close()

	kClient := createTestKibanaClient(server.URL)

	savedObject := kibanaeckv1alpha1.SavedObject{
		Body: `{"title": "My Dashboard"}`,
	}

	result, err := UpsertSavedObject(kClient, "dashboard", metav1.ObjectMeta{Name: "my-dashboard"}, savedObject)

	if err != nil {
		t.Errorf("UpsertSavedObject() unexpected error: %v", err)
	}

	if result != (ctrl.Result{}) {
		t.Errorf("UpsertSavedObject() result = %v, want empty Result", result)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 HTTP calls, got %d", callCount)
	}
}

func TestUpsertSavedObject_Update(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call - check existence (GET)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "existing-dashboard"}`))
			return
		}

		if callCount == 2 {
			// Second call - update (PUT)
			if r.Method != http.MethodPut {
				t.Errorf("Second call: expected PUT method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "existing-dashboard"}`))
			return
		}
	}))
	defer server.Close()

	kClient := createTestKibanaClient(server.URL)

	savedObject := kibanaeckv1alpha1.SavedObject{
		Body: `{"title": "Updated Dashboard"}`,
	}

	result, err := UpsertSavedObject(kClient, "dashboard", metav1.ObjectMeta{Name: "existing-dashboard"}, savedObject)

	if err != nil {
		t.Errorf("UpsertSavedObject() unexpected error: %v", err)
	}

	if result != (ctrl.Result{}) {
		t.Errorf("UpsertSavedObject() result = %v, want empty Result", result)
	}
}

func TestUpsertSavedObject_ServerError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"statusCode": 404}`))
			return
		}

		if callCount == 2 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"statusCode": 400, "error": "Bad Request"}`))
			return
		}
	}))
	defer server.Close()

	kClient := createTestKibanaClient(server.URL)

	savedObject := kibanaeckv1alpha1.SavedObject{
		Body: `{"invalid": "body"}`,
	}

	_, err := UpsertSavedObject(kClient, "dashboard", metav1.ObjectMeta{Name: "bad-dashboard"}, savedObject)

	if err == nil {
		t.Error("UpsertSavedObject() expected error for bad request, got nil")
	}
}

func TestDependenciesFulfilled(t *testing.T) {
	tests := []struct {
		name        string
		savedObject kibanaeckv1alpha1.SavedObject
		mockExists  map[string]bool // key: "type/name", value: exists
		wantErr     bool
	}{
		{
			name: "no dependencies",
			savedObject: kibanaeckv1alpha1.SavedObject{
				Dependencies: []kibanaeckv1alpha1.Dependency{},
			},
			mockExists: map[string]bool{},
			wantErr:    false,
		},
		{
			name: "all dependencies fulfilled",
			savedObject: kibanaeckv1alpha1.SavedObject{
				Dependencies: []kibanaeckv1alpha1.Dependency{
					{ObjectType: "visualization", Name: "viz-1"},
					{ObjectType: "index-pattern", Name: "logs-*"},
				},
			},
			mockExists: map[string]bool{
				"visualization/viz-1":  true,
				"index-pattern/logs-*": true,
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			savedObject: kibanaeckv1alpha1.SavedObject{
				Dependencies: []kibanaeckv1alpha1.Dependency{
					{ObjectType: "visualization", Name: "viz-1"},
					{ObjectType: "index-pattern", Name: "nonexistent"},
				},
			},
			mockExists: map[string]bool{
				"visualization/viz-1":       true,
				"index-pattern/nonexistent": false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract type and name from path
				path := r.URL.Path
				exists := false
				for key, val := range tt.mockExists {
					if contains(path, key) {
						exists = val
						break
					}
				}

				if exists {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"id": "test"}`))
				} else {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"statusCode": 404}`))
				}
			}))
			defer server.Close()

			kClient := createTestKibanaClient(server.URL)

			err := DependenciesFulfilled(kClient, tt.savedObject)

			if (err != nil) != tt.wantErr {
				t.Errorf("DependenciesFulfilled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to create a test Kibana client
func createTestKibanaClient(serverURL string) Client {
	return Client{
		Cli: nil,
		Ctx: nil,
		KibanaSpec: configv2.KibanaSpec{
			Url: serverURL,
		},
		Req: ctrl.Request{},
	}
}
