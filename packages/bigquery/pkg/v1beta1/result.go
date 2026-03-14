package v1beta1

import (
	"encoding/base64"
	"encoding/json"
	"math"

	"cloud.google.com/go/bigquery"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	QueryResults struct {
		pageSize  int
		prevPage  resultPage
		nextPage  resultPage
		totalRows int64
		rows      []*structpb.Struct
	}
	resultRow  map[string]any
	resultPage interface {
		encode() string
	}
	readRowFunc func(resultRow) error
)

func decodeResultPage[T resultPage](data string) (p *T, err error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	p = new(T)
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r resultRow) Row() (*structpb.Struct, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	rowStruct := new(structpb.Struct)
	err = protojson.Unmarshal(b, rowStruct)
	if err != nil {
		return nil, err
	}

	return rowStruct, nil
}

func (tr *QueryResults) Rows() []*structpb.Struct {
	return tr.rows
}

func (r *resultRow) Load(v []bigquery.Value, s bigquery.Schema) error {
	var row = map[string]any{}
	for idx, f := range s {
		row[f.Name] = v[idx]
	}
	*r = row
	return nil
}

func (tr *QueryResults) Load(v []bigquery.Value, s bigquery.Schema) error {
	var row = map[string]any{}
	for idx, f := range s {
		row[f.Name] = v[idx]
	}

	b, err := json.Marshal(row)
	if err != nil {
		return err
	}

	rowStruct := new(structpb.Struct)
	err = protojson.Unmarshal(b, rowStruct)
	if err != nil {
		return err
	}

	tr.rows = append(tr.rows, rowStruct)

	return nil
}

func (tr *QueryResults) PrevPage() *string {
	if tr.prevPage != nil {
		data := tr.prevPage.encode()
		return &data
	}
	return nil
}

func (tr *QueryResults) NextPage() *string {
	if tr.nextPage != nil {
		data := tr.nextPage.encode()
		return &data
	}
	return nil
}

func (tr *QueryResults) TotalPages() uint64 {
	if tr.pageSize == 0 {
		return 0
	}
	return uint64(math.Ceil(float64(tr.totalRows) / float64(tr.pageSize)))
}

func (tr *QueryResults) TotalRows() uint64 {
	return uint64(tr.totalRows)
}

func (tr *QueryResults) RowsCount() uint64 {
	return uint64(len(tr.rows))
}
