package v1alpha1

type CommonElasticsearchConfig struct {
	// +optional
	ElasticsearchInstance string `json:"name,omitempty"`
	// +optional
	ElasticsearchInstanceNamespace string `json:"namespace,omitempty"`
}

// UpdateMode defines how updates to the resource should be handled
// +kubebuilder:validation:Enum=Overwrite;Block
type UpdateMode string

const (
	// UpdateModeOverwrite allows updates to overwrite the existing resource
	UpdateModeOverwrite UpdateMode = "Overwrite"
	// UpdateModeBlock blocks updates to the resource after initial creation
	UpdateModeBlock UpdateMode = "Block"
)

// UpdatePolicySpec defines the policy for handling updates to the resource
type UpdatePolicySpec struct {
	// UpdateMode defines how updates should be handled. Defaults to Overwrite.
	// +kubebuilder:default=Overwrite
	// +optional
	UpdateMode UpdateMode `json:"updateMode,omitempty"`
}
