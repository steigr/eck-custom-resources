package utils

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetRegisteredGVKsInGroupWithTemplatingSpec returns all GroupVersionKinds registered in the scheme
// that belong to the specified API group and have a "Template" property in their Spec.
func GetRegisteredGVKsInGroupWithTemplatingSpec(scheme *runtime.Scheme, group string) []schema.GroupVersionKind {
	var gvks []schema.GroupVersionKind

	for gvk, typ := range scheme.AllKnownTypes() {
		if gvk.Group == group {
			// Skip List types
			if len(gvk.Kind) > 4 && gvk.Kind[len(gvk.Kind)-4:] == "List" {
				continue
			}

			// Check if the type has a Spec field with a Template property
			if hasTemplateInSpec(typ) {
				gvks = append(gvks, gvk)
			}
		}
	}

	return gvks
}

// hasTemplateInSpec checks if the given type has a Spec field that contains a Template property.
func hasTemplateInSpec(typ reflect.Type) bool {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return false
	}

	// Look for Spec field
	specField, found := typ.FieldByName("Spec")
	if !found {
		return false
	}

	specType := specField.Type
	if specType.Kind() == reflect.Ptr {
		specType = specType.Elem()
	}

	if specType.Kind() != reflect.Struct {
		return false
	}

	// Check if Spec has a Template field
	_, hasTemplate := specType.FieldByName("Template")
	return hasTemplate
}

// ListResourcesReferencingResourceTemplateData lists all resources of the given GVK
// that reference the specified ResourceTemplateData by name in their spec.template.references.
func ListResourcesReferencingResourceTemplateData(
	cli client.Client,
	ctx context.Context,
	gvk schema.GroupVersionKind,
	resourceTemplateDataName string,
	resourceTemplateDataNamespace string,
) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured

	// Create an unstructured list for the given GVK
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	})

	// List all resources of this type across all namespaces
	if err := cli.List(ctx, list); err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		if referencesResourceTemplateData(item, resourceTemplateDataName, resourceTemplateDataNamespace) {
			result = append(result, item)
		}
	}

	return result, nil
}

// referencesResourceTemplateData checks if the given unstructured resource references
// the specified ResourceTemplateData in its spec.template.references field.
func referencesResourceTemplateData(resource unstructured.Unstructured, resourceTemplateDataName string, resourceTemplateDataNamespace string) bool {
	// Get spec.template.references
	spec, found, err := unstructured.NestedMap(resource.Object, "spec")
	if err != nil || !found {
		return false
	}

	template, found, err := unstructured.NestedMap(spec, "template")
	if err != nil || !found {
		return false
	}

	references, found, err := unstructured.NestedSlice(template, "references")
	if err != nil || !found {
		return false
	}

	resourceNamespace := resource.GetNamespace()

	for _, ref := range references {
		refMap, ok := ref.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(refMap, "name")
		namespace, nsFound, _ := unstructured.NestedString(refMap, "namespace")

		// If namespace is not specified in the reference, use the resource's namespace
		if !nsFound || namespace == "" {
			namespace = resourceNamespace
		}

		// Check if this reference matches the ResourceTemplateData we're looking for
		if name == resourceTemplateDataName && namespace == resourceTemplateDataNamespace {
			return true
		}
	}

	return false
}
