package elasticsearch

import (
	"eck-custom-resources/utils"
	"encoding/json"
	"strings"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	ctrl "sigs.k8s.io/controller-runtime"
)

// IngestPipelineResponse represents the response from Elasticsearch Get Pipeline API
type IngestPipelineResponse struct {
	Description string           `json:"description,omitempty"`
	Processors  []map[string]any `json:"processors,omitempty"`
	OnFailure   []map[string]any `json:"on_failure,omitempty"`
	Version     int64            `json:"version,omitempty"`
	Meta        map[string]any   `json:"_meta,omitempty"`
}

func DeleteIngestPipeline(esClient *elasticsearch.Client, ingestPipelineId string) (ctrl.Result, error) {
	res, err := esClient.Ingest.DeletePipeline(ingestPipelineId)
	if err != nil || res.IsError() {
		return utils.GetRequeueResult(), err
	}
	return ctrl.Result{}, nil
}

func UpsertIngestPipeline(esClient *elasticsearch.Client, ingestPipeline v1alpha1.IngestPipeline, body string) (ctrl.Result, error) {
	res, err := esClient.Ingest.PutPipeline(ingestPipeline.Name, strings.NewReader(body))

	if err != nil || res.IsError() {
		return utils.GetRequeueResult(), GetClientErrorOrResponseError(err, res)
	}

	return ctrl.Result{}, nil
}

// GetIngestPipeline retrieves an ingest pipeline by ID from Elasticsearch
func GetIngestPipeline(esClient *elasticsearch.Client, pipelineId string) (*IngestPipelineResponse, error) {
	res, err := esClient.Ingest.GetPipeline(
		esClient.Ingest.GetPipeline.WithPipelineID(pipelineId),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, GetClientErrorOrResponseError(nil, res)
	}

	var pipelines map[string]IngestPipelineResponse
	if err := json.NewDecoder(res.Body).Decode(&pipelines); err != nil {
		return nil, err
	}

	pipeline, exists := pipelines[pipelineId]
	if !exists {
		return nil, nil
	}

	return &pipeline, nil
}
