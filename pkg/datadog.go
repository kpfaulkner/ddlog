package pkg

import (
	"encoding/json"
	"github.com/kpfaulkner/ddlog/pkg/comms"
	"github.com/kpfaulkner/ddlog/pkg/models"
	"time"
)

type Datadog struct {
  comms comms.DatadogComms
}

func NewDatadog(apiKey string, appKey string) *Datadog {
	d := Datadog{}
	d.comms = comms.NewDatadogComms(apiKey, appKey)
	return &d
}

// QueryDatadog does the query... duh :)
func (d *Datadog) QueryDatadog(query string, from time.Time, to time.Time) (*models.DatadogQueryResponse,error) {
	ddQuery := models.GenerateDatadogQuery(query, from, to)
	queryBytes,err := json.Marshal(ddQuery)
	if err != nil {
		return nil, err
	}
	ddResp, err := d.queryDatadogWithGeneratedQuery(queryBytes)
	//ddResp.NextLogID
	return ddResp, err
}

// QueryDatadogWithStartAt does the query but also uses the StartAt feature so will only return log entries from "startat" position onwards
func (d *Datadog) QueryDatadogWithStartAt(query string, from time.Time, to time.Time, startAt string) (*models.DatadogQueryResponse,error) {
	ddQuery := models.GenerateDatadogQueryWithStartAt(query, from, to, startAt)
	queryBytes,err := json.Marshal(ddQuery)
	if err != nil {
		return nil, err
	}
	ddResp, err := d.queryDatadogWithGeneratedQuery(queryBytes)
	return ddResp, err
}

// queryDatadog does the query.
func (d *Datadog) queryDatadogWithGeneratedQuery(queryBytes []byte) (*models.DatadogQueryResponse,error) {

	resp, err := d.comms.DoPost(queryBytes)
	if err != nil {
		return nil, err
	}

	var ddResp models.DatadogQueryResponse
	err = json.Unmarshal(resp, &ddResp)
	if err != nil {
		return nil, err
	}

	return &ddResp, err
}

