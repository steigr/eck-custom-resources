package template

import (
	"context"

	eseckv1alpha1 "eck-custom-resources/api/es.eck/v1alpha1"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchAndRenderTemplate fetches all referenced ResourceTemplateData objects and renders the body template.
// If the template spec has no references, it returns the original body unchanged.
// This function combines FetchResourceTemplateData and RenderBody for convenience.
func FetchAndRenderTemplate(
	cli client.Client,
	ctx context.Context,
	templateSpec eseckv1alpha1.CommonTemplatingSpec,
	body string,
	defaultNamespace string,
	restConfig *rest.Config,
) (string, error) {
	// If templating is not enabled or no references, return the original body
	if !IsTemplate(templateSpec) {
		return body, nil
	}

	// Fetch all referenced ResourceTemplateData objects
	resourceTemplateDataList, err := FetchResourceTemplateData(
		cli,
		ctx,
		templateSpec,
		defaultNamespace,
	)
	if err != nil {
		return "", err
	}

	// Render the body template with the fetched data
	return RenderBody(body, resourceTemplateDataList, restConfig)
}
