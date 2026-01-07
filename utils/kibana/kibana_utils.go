package kibana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	configv2 "eck-custom-resources/api/config/v2"
	kibanaeckv1alpha1 "eck-custom-resources/api/kibana.eck/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InjectId(objectJson string, id string) (*string, error) {
	var body map[string]interface{}
	err := json.NewDecoder(strings.NewReader(objectJson)).Decode(&body)
	if err != nil {
		return nil, err
	}

	body["id"] = id

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	sBody := string(marshalledBody)
	return &sBody, nil
}

func GetTargetInstance(cli client.Client, ctx context.Context, namespace string, targetName string, kibanaInstance *kibanaeckv1alpha1.KibanaInstance) error {
	if err := cli.Get(ctx, client.ObjectKey{Namespace: namespace, Name: targetName}, kibanaInstance); err != nil {
		return err
	}
	return nil
}

// GetKibanaTargetInstance resolves the target Kibana instance from either the project config
// or a named KibanaInstance resource. It returns the KibanaSpec to use for API calls.
func GetKibanaTargetInstance(
	cli client.Client,
	ctx context.Context,
	recorder record.EventRecorder,
	object runtime.Object,
	defaultKibana configv2.KibanaSpec,
	targetConfig kibanaeckv1alpha1.CommonKibanaConfig,
	namespace string,
) (*configv2.KibanaSpec, error) {
	targetInstance := defaultKibana
	if targetConfig.KibanaInstance != "" {
		if targetConfig.KibanaInstanceNamespace != "" {
			namespace = targetConfig.KibanaInstanceNamespace
		}
		var resourceInstance kibanaeckv1alpha1.KibanaInstance
		if err := GetTargetInstance(cli, ctx, namespace, targetConfig.KibanaInstance, &resourceInstance); err != nil {
			recorder.Event(object, "Warning", "Failed to load target instance", fmt.Sprintf("Target instance not found: %s", err.Error()))
			return nil, err
		}

		targetInstance = resourceInstance.Spec
	}
	return &targetInstance, nil
}
