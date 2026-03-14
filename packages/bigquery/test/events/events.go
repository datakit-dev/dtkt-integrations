package bigqueryevents

import (
	_ "embed"
	"encoding/json"
)

//go:embed dataset_events.json
var datasetEventsData []byte
var DatasetEvents []map[string]any

//go:embed query_events.json
var queryEventsData []byte
var QueryEvents []map[string]any

func init() {
	var err error
	err = json.Unmarshal(datasetEventsData, &DatasetEvents)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(queryEventsData, &QueryEvents)
	if err != nil {
		panic(err)
	}
}
