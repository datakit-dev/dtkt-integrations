package pkg

import (
	_ "embed"
	"encoding/json"
	"net/http"
)

//go:embed type/api_request_types.json
var reqTypesData []byte

//go:embed type/api_enum_types.json
var enumTypesData []byte

type (
	RequestTypes []RequestType
	RequestType  struct {
		Name   string            `json:"name"`
		Fields map[string]string `json:"fields"`
		Roots  []Root            `json:"root"`
		Edges  []Edge            `json:"edges"`
	}
	EnumTypes []EnumType
	EnumType  struct {
		Name         string   `json:"name"`
		Node         string   `json:"node"`
		FieldOrParam string   `json:"field_or_param"`
		Values       []string `json:"values"`
	}
	Root struct {
		Name   string        `json:"name"`
		Method RequestMethod `json:"method"`
		Return string        `json:"return"`
		Params []struct {
			Name     string `json:"name"`
			Required bool   `json:"required"`
			Type     string `json:"type"`
		} `json:"params"`
	}
	Edge struct {
		Method   RequestMethod `json:"method"`
		Endpoint string        `json:"endpoint"`
		Return   string        `json:"return"`
		Params   []struct {
			Name     string `json:"name"`
			Required bool   `json:"required"`
			Type     string `json:"type"`
		} `json:"params"`
	}
	RequestMethod string
)

const (
	GetRequest     = RequestMethod(http.MethodGet)
	HeadRequest    = RequestMethod(http.MethodHead)
	PostRequest    = RequestMethod(http.MethodPost)
	PutRequest     = RequestMethod(http.MethodPut)
	PatchRequest   = RequestMethod(http.MethodPatch)
	DeleteRequest  = RequestMethod(http.MethodDelete)
	ConnectRequest = RequestMethod(http.MethodConnect)
	OptionsRequest = RequestMethod(http.MethodOptions)
	TraceRequest   = RequestMethod(http.MethodTrace)
)

var RequestMethods = []RequestMethod{
	GetRequest,
	HeadRequest,
	PostRequest,
	PutRequest,
	PatchRequest,
	DeleteRequest,
	ConnectRequest,
	OptionsRequest,
	TraceRequest,
}

func (r RequestMethod) String() string {
	return string(r)
}

func LoadRequestTypes() (reqTypes RequestTypes, err error) {
	if err := json.Unmarshal(reqTypesData, &reqTypes); err != nil {
		return nil, err
	}
	return reqTypes, nil
}

func LoadEnumTypes() (enumTypes EnumTypes, err error) {
	if err := json.Unmarshal(enumTypesData, &enumTypes); err != nil {
		return nil, err
	}
	return enumTypes, nil
}
