package elasticsearch

import (
	"eck-custom-resources/utils"
	"strings"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	ctrl "sigs.k8s.io/controller-runtime"
)

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
