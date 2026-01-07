package template

import (
	"context"
	"fmt"

	eseckv1alpha1 "eck-custom-resources/api/es.eck/v1alpha1"

	"helm.sh/helm/v4/pkg/chart/common"
	"helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/engine"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const templateName = "body.tpl"

// FetchResourceTemplateData fetches all ResourceTemplateData objects referenced in the template spec.
// It handles both direct name references and label selector references.
func FetchResourceTemplateData(
	cli client.Client,
	ctx context.Context,
	templateSpec eseckv1alpha1.CommonTemplatingSpec,
	defaultNamespace string,
) ([]eseckv1alpha1.ResourceTemplateData, error) {
	var result []eseckv1alpha1.ResourceTemplateData
	seen := make(map[string]bool) // Track seen resources to avoid duplicates

	for _, ref := range templateSpec.References {
		namespace := defaultNamespace
		if ref.Namespace != "" {
			namespace = ref.Namespace
		}

		if ref.Name != "" {
			// Direct name reference
			var rtd eseckv1alpha1.ResourceTemplateData
			if err := cli.Get(ctx, client.ObjectKey{Namespace: namespace, Name: ref.Name}, &rtd); err != nil {
				return nil, fmt.Errorf("failed to get ResourceTemplateData %s/%s: %w", namespace, ref.Name, err)
			}
			key := namespace + "/" + rtd.Name
			if !seen[key] {
				result = append(result, rtd)
				seen[key] = true
			}
		}

		if len(ref.LabelSelector) > 0 {
			// Label selector reference
			var rtdList eseckv1alpha1.ResourceTemplateDataList
			listOpts := &client.ListOptions{
				Namespace:     namespace,
				LabelSelector: labels.SelectorFromSet(ref.LabelSelector),
			}
			if err := cli.List(ctx, &rtdList, listOpts); err != nil {
				return nil, fmt.Errorf("failed to list ResourceTemplateData with selector %v in namespace %s: %w", ref.LabelSelector, namespace, err)
			}
			for _, rtd := range rtdList.Items {
				key := rtd.Namespace + "/" + rtd.Name
				if !seen[key] {
					result = append(result, rtd)
					seen[key] = true
				}
			}
		}
	}

	return result, nil
}

// HasTemplateReferences checks if the template spec has any references defined.
func HasTemplateReferences(templateSpec eseckv1alpha1.CommonTemplatingSpec) bool {
	return len(templateSpec.References) > 0
}

// RenderBody renders the given body template using data from ResourceTemplateData objects.
// It uses the Helm template engine for rendering.
// The data from all ResourceTemplateData objects is merged into a single map,
// where each ResourceTemplateData's data is accessible via its name.
func RenderBody(body string, resourceTemplateDataList []eseckv1alpha1.ResourceTemplateData, config *rest.Config) (string, error) {
	// Build the template data map
	// Structure: { "resourceName": { "key1": "value1", "key2": "value2" }, ... }
	data := make(map[string]interface{})

	for _, rtd := range resourceTemplateDataList {
		// Convert map[string]string to map[string]interface{} for template compatibility
		rtdData := make(map[string]interface{})
		for k, v := range rtd.Data {
			rtdData[k] = v
		}
		data[rtd.Name] = rtdData
	}

	return RenderBodyWithValues(body, data, config)
}

// RenderBodyWithValues renders the given body template using a pre-built values map.
// This is useful when you want more control over the template data structure.
// Values are accessible in templates via .Values.key syntax (Helm convention).
func RenderBodyWithValues(body string, values map[string]interface{}, config *rest.Config) (string, error) {
	// Create a minimal chart with just our template
	chrt := &v2.Chart{
		Metadata: &v2.Metadata{
			Name:       "body-template",
			Version:    "0.0.0",
			APIVersion: v2.APIVersionV2,
		},
		Templates: []*common.File{
			{
				Name: "templates/" + templateName,
				Data: []byte(body),
			},
		},
	}

	// Wrap values under "Values" key as Helm expects
	wrappedValues := map[string]interface{}{
		"Values": values,
	}

	// Render the chart using RenderWithClient to enable client-aware template functions (e.g., lookup)
	rendered, err := engine.RenderWithClient(chrt, wrappedValues, config)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	// Get the rendered template by its full path
	fullTemplateName := "body-template/templates/" + templateName
	result, ok := rendered[fullTemplateName]
	if !ok {
		return "", fmt.Errorf("template %s not found in rendered output", fullTemplateName)
	}

	return result, nil
}
