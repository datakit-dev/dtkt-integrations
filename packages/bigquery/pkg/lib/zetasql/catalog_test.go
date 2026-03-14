package zetasql_test

import (
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib/zetasql"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
)

func sharedCatalog1() *catalogv1beta1.CatalogPermission {
	return v1beta1.NewCatalogPermission(
		"datakit-bigquery",
		"shared_bigquery1",
		v1beta1.NewSchemaPermission(
			"shared_schema1",
			"shared_schema1",
			v1beta1.NewTablePermission(
				"shared_table1",
				"shared_table1",
				v1beta1.NewColumnPermission(
					"shared_column1",
					"shared_column1",
					string(bigquery.StringFieldType),
				),
			),
		),
	)
}

func sharedCatalog2() *catalogv1beta1.CatalogPermission {
	return v1beta1.NewCatalogPermission(
		"datakit-bigquery",
		"shared_bigquery2",
		v1beta1.NewSchemaPermission(
			"shared_schema2",
			"shared_schema2",
			v1beta1.NewTablePermission(
				"shared_table2",
				"shared_table2",
				v1beta1.NewColumnPermission(
					"shared_column2",
					"shared_column2",
					string(bigquery.StringFieldType),
				),
			),
		),
	)
}

func publicCatalog() *catalogv1beta1.CatalogPermission {
	return v1beta1.NewCatalogPermission(
		"bigquery-public-data",
		"bigquery_public",
		v1beta1.NewSchemaPermission(
			"public_schema1",
			"public_schema1",
			v1beta1.NewTablePermission(
				"public_table1",
				"public_table1",
				v1beta1.NewColumnPermission(
					"public_column1",
					"public_column1",
					string(bigquery.StringFieldType),
				),
				v1beta1.NewColumnPermission(
					"public_column2",
					"public_column2",
					string(bigquery.IntegerFieldType),
				),
			),
		),
	)
}

func TestRootCatalog_SingleCatalog(t *testing.T) {
	root, err := zetasql.NewRootCatalog([]*catalogv1beta1.CatalogPermission{sharedCatalog1()})
	if err != nil {
		t.Fatal(err)
	}

	tbl, err := root.FindTable([]string{"datakit-bigquery", "shared_schema1", "shared_table1"})
	if err != nil {
		t.Fatal(err)
	}

	if tbl == nil {
		t.Fatal("table not found")
	} else if tbl.Name() != "shared_table1" {
		t.Fatalf("unexpected table name: %s", tbl.Name())
	}
}

func TestRootCatalog_MultipleCatalogs_DifferentName(t *testing.T) {
	root, err := zetasql.NewRootCatalog([]*catalogv1beta1.CatalogPermission{
		sharedCatalog1(),
		publicCatalog(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tbl1, err := root.FindTable([]string{"datakit-bigquery", "shared_schema1", "shared_table1"})
	if err != nil {
		t.Fatal(err)
	}

	if tbl1 == nil {
		t.Fatal("table not found")
	} else if tbl1.Name() != "shared_table1" {
		t.Fatalf("unexpected table name: %s", tbl1.Name())
	}

	tbl2, err := root.FindTable([]string{"bigquery-public-data", "public_schema1", "public_table1"})
	if err != nil {
		t.Fatal(err)
	}

	if tbl2 == nil {
		t.Fatal("table not found")
	}

	if tbl2.Name() != "public_table1" {
		t.Fatalf("unexpected table name: %s", tbl2.Name())
	}
}

func TestRootCatalog_MultipleCatalogs_SameName(t *testing.T) {
	root, err := zetasql.NewRootCatalog([]*catalogv1beta1.CatalogPermission{
		sharedCatalog1(),
		sharedCatalog2(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tbl1, err := root.FindTable([]string{"datakit-bigquery", "shared_schema1", "shared_table1"})
	if err != nil {
		t.Fatal(err)
	}

	if tbl1 == nil {
		t.Fatal("table not found")
	} else if tbl1.Name() != "shared_table1" {
		t.Fatalf("unexpected table name: %s", tbl1.Name())
	}

	tbl2, err := root.FindTable([]string{"datakit-bigquery", "shared_schema2", "shared_table2"})
	if err != nil {
		t.Fatal(err)
	}

	if tbl2 == nil {
		t.Fatal("table not found")
	}

	if tbl2.Name() != "shared_table2" {
		t.Fatalf("unexpected table name: %s", tbl2.Name())
	}
}
