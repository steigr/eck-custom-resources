package v1alpha1

type CommonElasticsearchConfig struct {
	// +optional
	ElasticsearchInstance string `json:"name,omitempty"`
	// +optional
	ElasticsearchInstanceNamespace string `json:"namespace,omitempty"`
}
