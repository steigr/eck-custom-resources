package v1alpha1

type CommonTemplatingSpec struct {
	// +optional
	References []CommonTemplatingSpecReference `json:"references,omitempty"`
}

type CommonTemplatingSpecReference struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// labelSeclector to select ResourceTemplateData objects
	// +optional
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
}
