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

func TestSpaceExists(t *testing.T) {
	tests := []struct {
		name             string
		spaceName        string
		serverStatusCode int
		serverResponse   string
		wantExists       bool
		wantErr          bool
	}{
		{
			name:             "space exists",
			spaceName:        "my-space",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"id": "my-space", "name": "My Space"}`,
			wantExists:       true,
			wantErr:          false,
		},
		{
			name:             "space does not exist",
			spaceName:        "nonexistent",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"statusCode": 404, "error": "Not Found"}`,
			wantExists:       false,
			wantErr:          false,
		},
		{
			name:             "server error",
			spaceName:        "my-space",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": "Internal Server Error"}`,
			wantExists:       false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/spaces/space/" + tt.spaceName
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

			kClient := createTestClient(server.URL)

			exists, err := SpaceExists(kClient, tt.spaceName)

			if (err != nil) != tt.wantErr {
				t.Errorf("SpaceExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("SpaceExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestDeleteSpace(t *testing.T) {
	tests := []struct {
		name             string
		spaceName        string
		serverStatusCode int
		serverResponse   string
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			spaceName:        "my-space",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{}`,
			wantErr:          false,
		},
		{
			name:             "space not found",
			spaceName:        "nonexistent",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"statusCode": 404}`,
			wantErr:          false, // DeleteSpace doesn't return error on 404
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/spaces/space/" + tt.spaceName
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE method, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			kClient := createTestClient(server.URL)

			result, err := DeleteSpace(kClient, tt.spaceName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSpace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify result is empty (no requeue)
			if !tt.wantErr && result != (ctrl.Result{}) {
				t.Errorf("DeleteSpace() result = %v, want empty Result", result)
			}
		})
	}
}

func TestUpsertSpace_Create(t *testing.T) {
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
			if r.URL.Path != "/api/spaces/space" {
				t.Errorf("Expected path /api/spaces/space, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "new-space"}`))
			return
		}
	}))
	defer server.Close()

	kClient := createTestClient(server.URL)

	space := kibanaeckv1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-space",
			Namespace: "default",
		},
		Spec: kibanaeckv1alpha1.SpaceSpec{
			Body: `{"name": "New Space", "description": "A new space"}`,
		},
	}

	result, err := UpsertSpace(kClient, space)

	if err != nil {
		t.Errorf("UpsertSpace() unexpected error: %v", err)
	}

	if result != (ctrl.Result{}) {
		t.Errorf("UpsertSpace() result = %v, want empty Result", result)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 HTTP calls, got %d", callCount)
	}
}

func TestUpsertSpace_Update(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call - check existence (GET)
			if r.Method != http.MethodGet {
				t.Errorf("First call: expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "existing-space"}`))
			return
		}

		if callCount == 2 {
			// Second call - update (PUT)
			if r.Method != http.MethodPut {
				t.Errorf("Second call: expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/api/spaces/space/existing-space" {
				t.Errorf("Expected path /api/spaces/space/existing-space, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "existing-space"}`))
			return
		}
	}))
	defer server.Close()

	kClient := createTestClient(server.URL)

	space := kibanaeckv1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-space",
			Namespace: "default",
		},
		Spec: kibanaeckv1alpha1.SpaceSpec{
			Body: `{"name": "Existing Space", "description": "Updated description"}`,
		},
	}

	result, err := UpsertSpace(kClient, space)

	if err != nil {
		t.Errorf("UpsertSpace() unexpected error: %v", err)
	}

	if result != (ctrl.Result{}) {
		t.Errorf("UpsertSpace() result = %v, want empty Result", result)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 HTTP calls, got %d", callCount)
	}
}

func TestUpsertSpace_ErrorResponse(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call - check existence
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"statusCode": 404}`))
			return
		}

		if callCount == 2 {
			// Second call - create fails
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"statusCode": 400, "error": "Bad Request", "message": "Invalid space configuration"}`))
			return
		}
	}))
	defer server.Close()

	kClient := createTestClient(server.URL)

	space := kibanaeckv1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bad-space",
			Namespace: "default",
		},
		Spec: kibanaeckv1alpha1.SpaceSpec{
			Body: `{"invalid": "config"}`,
		},
	}

	_, err := UpsertSpace(kClient, space)

	if err == nil {
		t.Error("UpsertSpace() expected error for bad request, got nil")
	}
}

// Helper function to create a test Kibana client
func createTestClient(serverURL string) Client {
	return Client{
		Cli: nil, // Not needed for these tests
		Ctx: nil, // Not needed for these tests
		KibanaSpec: configv2.KibanaSpec{
			Url: serverURL,
		},
		Req: ctrl.Request{},
	}
}
