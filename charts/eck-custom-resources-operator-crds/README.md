# eck-custom-resources-operator-crds

A Helm chart for installing the Custom Resource Definitions (CRDs) required by the eck-custom-resources operator.

## Description

This chart contains the CRDs for the eck-custom-resources operator. It is designed to be installed separately from the operator itself, which allows for:

- Independent lifecycle management of CRDs
- Pre-installation of CRDs before the operator
- Easier upgrades without accidentally deleting CRDs

## Installation

```bash
helm install eck-custom-resources-crds ./charts/eck-custom-resources-operator-crds
```

## CRDs Included

### Elasticsearch CRDs (es.eck.github.com)

- ComponentTemplate
- ElasticsearchApiKey
- ElasticsearchInstance
- ElasticsearchRole
- ElasticsearchUser
- Index
- IndexLifecyclePolicy
- IndexTemplate
- IngestPipeline
- ResourceTemplateData
- SnapshotLifecyclePolicy
- SnapshotRepository

### Kibana CRDs (kibana.eck.github.com)

- Dashboard
- DataView
- IndexPattern
- KibanaInstance
- Lens
- SavedSearch
- Space
- Visualization

### Config CRDs (config.github.com)

- ProjectConfig

## Uninstallation

```bash
helm uninstall eck-custom-resources-crds
```

**Warning:** Uninstalling this chart will delete the CRDs and all custom resources of these types.
