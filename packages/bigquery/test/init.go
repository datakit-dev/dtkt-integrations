package test

import (
	"bytes"
	_ "embed"
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/goccy/bigquery-emulator/types"
	"github.com/goccy/go-yaml"
)

//go:embed emulator.yaml
var emulator []byte
var BigQueryProjects []*types.Project

func init() {
	// Load emulator.yaml
	validate := validator.New()
	types.RegisterTypeValidation(validate)
	dec := yaml.NewDecoder(
		bytes.NewBuffer(emulator),
		yaml.Validator(validate),
		yaml.Strict(),
	)
	var v struct {
		Projects []*types.Project `yaml:"projects" validate:"required"`
	}
	if err := dec.Decode(&v); err != nil {
		panic(errors.New(yaml.FormatError(err, false, true)))
	}

	BigQueryProjects = v.Projects
}
