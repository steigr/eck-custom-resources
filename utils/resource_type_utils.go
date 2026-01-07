package utils

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
// It also checks that the dependent resource's targetInstance namespace matches.
func ListResourcesReferencingResourceTemplateData(
	cli client.Client,
	ctx context.Context,
	gvk schema.GroupVersionKind,
	resourceTemplateDataName string,
	resourceTemplateDataNamespace string,
) ([]unstructured.Unstructured, error) {
	logger := log.FromContext(ctx)
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
		logger.Info("Checking resource for ResourceTemplateData reference", "GVK", gvk, "Name", item.GetName(), "Namespace", item.GetNamespace())
		if referencesResourceTemplateDataItemsAndHasSameTargetInstance(item, resourceTemplateDataName, resourceTemplateDataNamespace) {
			logger.Info("Resource references ResourceTemplateData and has matching targetInstance", "GVK", gvk, "Name", item.GetName(), "Namespace", item.GetNamespace())
			result = append(result, item)
		}
	}

	return result, nil
}

// referencesResourceTemplateDataItemsAndHasSameTargetInstance checks if the given unstructured resource has
// the spec.template.references field present and matches the specified targetInstance name and namespace (if given).
func referencesResourceTemplateDataItemsAndHasSameTargetInstance(resource unstructured.Unstructured, targetInstanceName string, targetInstanceNamespace string) bool {
	// Get spec
	logger := log.FromContext(context.TODO())
	logger.Info("Checking if resource has spec.template.references and matching targetInstance", "Resource", resource.GetName(), "Namespace", resource.GetNamespace())
	spec, found, err := unstructured.NestedMap(resource.Object, "spec")
	if err != nil || !found {
		return false
	}

	// Check the resource's targetInstance matches the given targetInstance name and namespace
	if targetInstance, found, _ := unstructured.NestedMap(spec, "targetInstance"); found {
		logger.Info("Found spec.targetInstance in resource", "Resource", resource.GetName(), "Namespace", resource.GetNamespace())
		// Check targetInstance name if provided
		if targetInstanceName != "" {
			name, _, _ := unstructured.NestedString(targetInstance, "name")
			if name != targetInstanceName {
				return false
			}
			logger.Info("Matched targetInstance name", "Resource", resource.GetName(), "Namespace", resource.GetNamespace())
		}
		// Check targetInstance namespace if provided
		if targetInstanceNamespace != "" {
			namespace, _, _ := unstructured.NestedString(targetInstance, "namespace")
			// If namespace is set in targetInstance, it must match
			// If namespace is not set in targetInstance, the resource's namespace is used implicitly
			if namespace != "" && namespace != targetInstanceNamespace {
				return false
			}
			logger.Info("Matched targetInstance namespace", "Resource", resource.GetName(), "Namespace", resource.GetNamespace())
		}
	} else {
		// No targetInstance specified in resource - use resource's namespace for comparison
		if targetInstanceNamespace != "" && resource.GetNamespace() != targetInstanceNamespace {
			return false
		}
	}

	// Get spec.template.references - must exist for this to be a templated resource
	template, found, err := unstructured.NestedMap(spec, "template")
	if err != nil || !found {
		return false
	}
	logger.Info("Found spec.template in resource", "Resource", resource.GetName(), "Namespace", resource.GetNamespace())

	references, found, err := unstructured.NestedSlice(template, "references")
	if err != nil || !found || len(references) == 0 {
		return false
	}
	logger.Info("Found spec.template.references in resource", "Resource", resource.GetName(), "Namespace", resource.GetNamespace())

	return true
}
