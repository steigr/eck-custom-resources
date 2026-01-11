package kibana

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv2 "eck-custom-resources/api/config/v2"
	kibanaeckv1alpha1 "eck-custom-resources/api/kibana.eck/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestRemoveName(t *testing.T) {
	tests := []struct {
		name       string
		objectJson string
		id         string
		wantErr    bool
		wantName   bool // whether name should exist in result
	}{
		{
			name:       "remove name from object",
			objectJson: `{"name": "test-name", "title": "Test Title"}`,
			id:         "test-id",
			wantErr:    false,
			wantName:   false,
		},
		{
			name:       "object without name field",
			objectJson: `{"title": "Test Title", "description": "desc"}`,
			id:         "test-id",
			wantErr:    false,
			wantName:   false,
		},
		{
			name:       "empty object",
			objectJson: `{}`,
			id:         "test-id",
			wantErr:    false,
			wantName:   false,
		},
		{
			name:       "invalid json returns error",
			objectJson: `{invalid}`,
			id:         "test-id",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := removeName(tt.objectJson, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("removeName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if result == nil {
				t.Error("removeName() returned nil for valid input")
				return
			}

			// Parse result to check for name field
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(*result), &parsed); err != nil {
				t.Errorf("removeName() returned invalid JSON: %v", err)
				return
			}

			_, hasName := parsed["name"]
			if hasName != tt.wantName {
				t.Errorf("removeName() name field presence = %v, want %v", hasName, tt.wantName)
			}
		})
	}
}

func TestRemoveName_PreservesOtherFields(t *testing.T) {
	objectJson := `{"name": "to-remove", "title": "Test Title", "description": "desc"}`

	result, err := removeName(objectJson, "test-id")
	if err != nil {
		t.Fatalf("removeName() unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(*result), &parsed); err != nil {
		t.Fatalf("removeName() returned invalid JSON: %v", err)
	}

	// Verify other fields are preserved
	if parsed["title"] != "Test Title" {
		t.Error("removeName() did not preserve 'title' field")
	}
	if parsed["description"] != "desc" {
		t.Error("removeName() did not preserve 'description' field")
	}
}

func TestFormatDataViewUrl(t *testing.T) {
	tests := []struct {
		name     string
		space    *string
		expected string
	}{
		{
			name:     "nil space",
			space:    nil,
			expected: "/api/data_views/data_view",
		},
		{
			name:     "with space",
			space:    strPtr("my-space"),
			expected: "/s/my-space/api/data_views/data_view",
		},
		{
			name:     "with default space",
			space:    strPtr("default"),
			expected: "/s/default/api/data_views/data_view",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDataViewUrl(tt.space)
			if result != tt.expected {
				t.Errorf("formatDataViewUrl() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatExistingDataViewUrl(t *testing.T) {
	tests := []struct {
		name     string
		viewName string
		space    *string
		expected string
	}{
		{
			name:     "nil space",
			viewName: "my-dataview",
			space:    nil,
			expected: "/api/data_views/data_view/my-dataview",
		},
		{
			name:     "with space",
			viewName: "my-dataview",
			space:    strPtr("my-space"),
			expected: "/s/my-space/api/data_views/data_view/my-dataview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatExistingDataViewUrl(tt.viewName, tt.space)
			if result != tt.expected {
				t.Errorf("formatExistingDataViewUrl() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestWrapDataView_Create(t *testing.T) {
	// Test wrapping for create operation (isUpdate = false)
	dataView := createTestDataView("test-view", `{"title": "Test Pattern"}`, nil)

	result, err := wrapDataView(dataView, false)
	if err != nil {
		t.Fatalf("wrapDataView() unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(*result), &parsed); err != nil {
		t.Fatalf("wrapDataView() returned invalid JSON: %v", err)
	}

	// Check for data_view wrapper
	if _, ok := parsed["data_view"]; !ok {
		t.Error("wrapDataView() missing 'data_view' wrapper")
	}

	// Check for override flag (should be present for create)
	if override, ok := parsed["override"]; !ok || override != true {
		t.Error("wrapDataView() missing or incorrect 'override' flag for create")
	}

	// Check that refresh_fields is NOT present for create
	if _, ok := parsed["refresh_fields"]; ok {
		t.Error("wrapDataView() should not have 'refresh_fields' for create")
	}

	// Check that id was injected
	dataViewMap := parsed["data_view"].(map[string]interface{})
	if dataViewMap["id"] != "test-view" {
		t.Errorf("wrapDataView() id = %v, want 'test-view'", dataViewMap["id"])
	}
}

func TestWrapDataView_Update(t *testing.T) {
	// Test wrapping for update operation (isUpdate = true)
	dataView := createTestDataView("test-view", `{"title": "Test Pattern", "name": "should-be-removed"}`, nil)

	result, err := wrapDataView(dataView, true)
	if err != nil {
		t.Fatalf("wrapDataView() unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(*result), &parsed); err != nil {
		t.Fatalf("wrapDataView() returned invalid JSON: %v", err)
	}

	// Check for data_view wrapper
	if _, ok := parsed["data_view"]; !ok {
		t.Error("wrapDataView() missing 'data_view' wrapper")
	}

	// Check for refresh_fields flag (should be present for update)
	if refreshFields, ok := parsed["refresh_fields"]; !ok || refreshFields != true {
		t.Error("wrapDataView() missing or incorrect 'refresh_fields' flag for update")
	}

	// Check that override is NOT present for update
	if _, ok := parsed["override"]; ok {
		t.Error("wrapDataView() should not have 'override' for update")
	}

	// Check that name was removed
	dataViewMap := parsed["data_view"].(map[string]interface{})
	if _, hasName := dataViewMap["name"]; hasName {
		t.Error("wrapDataView() should have removed 'name' field for update")
	}
}

func TestWrapDataView_InvalidJson(t *testing.T) {
	dataView := createTestDataView("test-view", `{invalid json}`, nil)

	_, err := wrapDataView(dataView, false)
	if err == nil {
		t.Error("wrapDataView() expected error for invalid JSON, got nil")
	}
}

func TestDataViewExists(t *testing.T) {
	tests := []struct {
		name             string
		dataViewName     string
		space            *string
		serverStatusCode int
		serverResponse   string
		wantExists       bool
		wantErr          bool
	}{
		{
			name:             "data view exists - no space",
			dataViewName:     "my-dataview",
			space:            nil,
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"data_view": {"id": "my-dataview"}}`,
			wantExists:       true,
			wantErr:          false,
		},
		{
			name:             "data view exists - with space",
			dataViewName:     "my-dataview",
			space:            strPtr("my-space"),
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"data_view": {"id": "my-dataview"}}`,
			wantExists:       true,
			wantErr:          false,
		},
		{
			name:             "data view not found",
			dataViewName:     "nonexistent",
			space:            nil,
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"statusCode": 404}`,
			wantExists:       false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := formatExistingDataViewUrl(tt.dataViewName, tt.space)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			kClient := createDataViewTestClient(server.URL)
			dataView := createTestDataView(tt.dataViewName, `{"title": "test"}`, tt.space)

			exists, err := DataViewExists(kClient, dataView)

			if (err != nil) != tt.wantErr {
				t.Errorf("DataViewExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("DataViewExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestDeleteDataView(t *testing.T) {
	tests := []struct {
		name             string
		dataViewName     string
		space            *string
		serverStatusCode int
		wantErr          bool
	}{
		{
			name:             "successful deletion - no space",
			dataViewName:     "my-dataview",
			space:            nil,
			serverStatusCode: http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "successful deletion - with space",
			dataViewName:     "my-dataview",
			space:            strPtr("my-space"),
			serverStatusCode: http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "not found",
			dataViewName:     "nonexistent",
			space:            nil,
			serverStatusCode: http.StatusNotFound,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := formatExistingDataViewUrl(tt.dataViewName, tt.space)
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

			kClient := createDataViewTestClient(server.URL)
			dataView := createTestDataView(tt.dataViewName, `{"title": "test"}`, tt.space)

			result, err := DeleteDataView(kClient, dataView)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteDataView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != (ctrl.Result{}) {
				t.Errorf("DeleteDataView() result = %v, want empty Result", result)
			}
		})
	}
}

func TestUpsertDataView_Create(t *testing.T) {
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
			// Second call - create (POST to base URL)
			if r.Method != http.MethodPost {
				t.Errorf("Second call: expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/api/data_views/data_view" {
				t.Errorf("Expected path /api/data_views/data_view, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data_view": {"id": "new-dataview"}}`))
			return
		}
	}))
	defer server.Close()

	kClient := createDataViewTestClient(server.URL)
	dataView := createTestDataView("new-dataview", `{"title": "New Data View"}`, nil)

	result, err := UpsertDataView(kClient, dataView)

	if err != nil {
		t.Errorf("UpsertDataView() unexpected error: %v", err)
	}

	if result != (ctrl.Result{}) {
		t.Errorf("UpsertDataView() result = %v, want empty Result", result)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 HTTP calls, got %d", callCount)
	}
}

func TestUpsertDataView_Update(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call - check existence (GET)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data_view": {"id": "existing-dataview"}}`))
			return
		}

		if callCount == 2 {
			// Second call - update (POST to specific URL)
			if r.Method != http.MethodPost {
				t.Errorf("Second call: expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/api/data_views/data_view/existing-dataview" {
				t.Errorf("Expected path /api/data_views/data_view/existing-dataview, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data_view": {"id": "existing-dataview"}}`))
			return
		}
	}))
	defer server.Close()

	kClient := createDataViewTestClient(server.URL)
	dataView := createTestDataView("existing-dataview", `{"title": "Updated Data View"}`, nil)

	result, err := UpsertDataView(kClient, dataView)

	if err != nil {
		t.Errorf("UpsertDataView() unexpected error: %v", err)
	}

	if result != (ctrl.Result{}) {
		t.Errorf("UpsertDataView() result = %v, want empty Result", result)
	}
}

func TestUpsertDataView_ServerError(t *testing.T) {
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

	kClient := createDataViewTestClient(server.URL)
	dataView := createTestDataView("bad-dataview", `{"invalid": "body"}`, nil)

	_, err := UpsertDataView(kClient, dataView)

	if err == nil {
		t.Error("UpsertDataView() expected error for bad request, got nil")
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func createTestDataView(name string, body string, space *string) kibanaeckv1alpha1.DataView {
	return kibanaeckv1alpha1.DataView{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: kibanaeckv1alpha1.DataViewSpec{
			SavedObject: kibanaeckv1alpha1.SavedObject{
				Body:  body,
				Space: space,
			},
		},
	}
}

// Helper function to create a test Kibana client for data view tests
func createDataViewTestClient(serverURL string) Client {
	return Client{
		Cli: nil,
		Ctx: nil,
		KibanaSpec: configv2.KibanaSpec{
			Url: serverURL,
		},
		Req: ctrl.Request{},
	}
}
