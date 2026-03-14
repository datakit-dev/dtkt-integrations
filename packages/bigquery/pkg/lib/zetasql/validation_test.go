package zetasql_test

import (
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib/zetasql"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
)

func TestValidation_SingleCatalog(t *testing.T) {
	var (
		cat = sharedCatalog1()
		sch = cat.Schemas[0]
		tbl = sch.Tables[0]
		col = tbl.Columns[0]
	)

	v, err := zetasql.
		NewValidation(fmt.Sprintf("SELECT %s FROM %s.%s.%s", col.Name, cat.Name, sch.Name, tbl.Name), nil).
		ValidateQuery(cat)
	if err != nil {
		t.Fatal(err)
	}

	if len(v.Accessed()) == 0 {
		t.Fatalf("expected 1 catalog, got 0")
	}

	var (
		catA = v.Accessed()[0]
		schA = catA.Schemas[0]
		tblA = schA.Tables[0]
		colA = tblA.Columns[0]
	)

	if catA.Name != cat.Name {
		t.Fatalf("expected %s, got %s", cat.Name, catA.Name)
	} else if schA.Name != sch.Name {
		t.Fatalf("expected %s, got %s", sch.Name, schA.Name)
	} else if tblA.Name != tbl.Name {
		t.Fatalf("expected %s, got %s", tbl.Name, tblA.Name)
	} else if colA.Name != col.Name {
		t.Fatalf("expected %s, got %s", col.Name, colA.Name)
	}

	cat = sharedCatalog1()
	v, err = zetasql.
		NewValidation(fmt.Sprintf("SELECT %s FROM %s.%s.%s", col.Alias, cat.Alias, sch.Alias, tbl.Alias), nil).
		ValidateQuery(cat)
	if err != nil {
		t.Fatal(err)
	}

	if len(v.Accessed()) == 0 {
		t.Fatalf("expected 1 catalog, got 0")
	}

	catA = v.Accessed()[0]
	schA = catA.Schemas[0]
	tblA = schA.Tables[0]
	colA = tblA.Columns[0]

	if catA.Name != cat.Name {
		t.Fatalf("expected %s, got %s", cat.Name, catA.Name)
	} else if schA.Name != sch.Name {
		t.Fatalf("expected %s, got %s", sch.Name, schA.Name)
	} else if tblA.Name != tbl.Name {
		t.Fatalf("expected %s, got %s", tbl.Name, tblA.Name)
	} else if colA.Name != col.Name {
		t.Fatalf("expected %s, got %s", col.Name, colA.Name)
	}
}

func TestValidation_MultipleCatalogs(t *testing.T) {
	cat1 := sharedCatalog1()
	cat2 := publicCatalog()

	v, err := zetasql.
		NewValidation(`
		SELECT shared_column1, public_column1 FROM shared_schema1.shared_table1 AS s
		JOIN public_schema1.public_table1 AS p ON s.shared_column1 = p.public_column1
	`, nil).
		ValidateQuery(cat1, cat2)
	if err != nil {
		t.Fatal(err)
	}

	if len(v.Accessed()) != 2 {
		t.Fatalf("expected 2 catalogs, got %d", len(v.Accessed()))
	}

	var (
		totalSchemas, totalTables, totalColumns int
	)
	for _, cat := range v.Accessed() {
		if cat.Name != cat1.Name && cat.Name != cat2.Name {
			t.Fatalf("unexpected catalog: %s", cat.Name)
		}

		for _, sch := range cat.Schemas {
			totalSchemas++

			if sch.Name != "shared_schema1" && sch.Name != "public_schema1" {
				t.Fatalf("unexpected schema: %s", sch.Name)
			}
			for _, tbl := range sch.Tables {
				totalTables++

				if tbl.Name != "shared_table1" && tbl.Name != "public_table1" {
					t.Fatalf("unexpected table: %s", tbl.Name)
				}

				for _, col := range tbl.Columns {
					totalColumns++

					if col.Name != "shared_column1" && col.Name != "public_column1" {
						t.Fatalf("unexpected column: %s", col.Name)
					}
				}
			}
		}
	}

	if totalSchemas != 2 {
		t.Fatalf("expected 2 schemas, got %d", totalSchemas)
	}
	if totalTables != 2 {
		t.Fatalf("expected 2 tables, got %d", totalTables)
	}
	if totalColumns != 2 {
		t.Fatalf("expected 2 columns, got %d", totalColumns)
	}
}

func TestValidation_MultipleCatalogs_SameName(t *testing.T) {
	cat1 := sharedCatalog1()
	cat2 := publicCatalog()

	v, err := zetasql.
		NewValidation("SELECT shared_column1 FROM datakit-bigquery.shared_schema1.shared_table1", nil).
		ValidateQuery(cat1, cat2)
	if err != nil {
		t.Fatal(err)
	}

	if len(v.Accessed()) != 1 {
		t.Fatalf("expected 1 catalog, got %d", len(v.Accessed()))
	}

	cat := v.Accessed()[0]
	if cat.Name != cat1.Name {
		t.Fatalf("expected %s, got %s", cat1.Name, cat.Name)
	}

	if len(cat.Schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(cat.Schemas))
	}

	sch := cat.Schemas[0]
	if sch.Name != "shared_schema1" {
		t.Fatalf("expected shared_schema1, got %s", sch.Name)
	}

	if len(sch.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(sch.Tables))
	}

	tbl := sch.Tables[0]
	if tbl.Name != "shared_table1" {
		t.Fatalf("expected shared_table1, got %s", tbl.Name)
	}

	if len(tbl.Columns) != 1 {
		t.Fatalf("expected 1 column, got %d", len(tbl.Columns))
	}

	col := tbl.Columns[0]
	if col.Name != "shared_column1" {
		t.Fatalf("expected shared_column1, got %s", col.Name)
	}
}

func TestValidation_ScalarFieldTypes(t *testing.T) {
	var scalarTypes = map[bigquery.FieldType]any{
		bigquery.BooleanFieldType:    true,
		bigquery.StringFieldType:     "foo",
		bigquery.BytesFieldType:      []byte("foo"),
		bigquery.IntegerFieldType:    123,
		bigquery.NumericFieldType:    "456.78",
		bigquery.BigNumericFieldType: "56789.012",
		bigquery.DateFieldType:       time.Now().Format("2006-01-02"),
		bigquery.DateTimeFieldType:   time.Now().Format("2006-01-02 15:04:05"),
		bigquery.TimeFieldType:       time.Now().Format("15:04:05"),
		bigquery.TimestampFieldType:  time.Now().Format("2006-01-02 15:04:05"),
		bigquery.FloatFieldType:      123.45,
	}

	for fType, value := range scalarTypes {
		param, err := v1beta1.NewParamWithValue(
			v1beta1.NewField("scalar", v1beta1.NewDataType(fType)),
			value,
		)
		if err != nil {
			t.Fatal(err)
		}

		v, err := zetasql.
			NewValidation("SELECT @scalar AS scalar", v1beta1.NewParams(param)).
			ValidateQuery()
		if err != nil {
			t.Fatal(fmt.Errorf("scalar %s => %s error: %w", fType, zetasql.ZetaTypes[fType].Kind(), err))
		}

		for _, f := range v.Fields() {
			if f.Type.NativeType != string(fType) {
				t.Fatalf("expected %s, got %s", fType, f.Type)
			}
		}

		param, err = v1beta1.NewParamWithValue(
			v1beta1.NewField(
				"scalar_repeated",
				v1beta1.NewDataType(fType),
				v1beta1.WithFieldRepeated(true),
			),
			[]any{value},
		)
		if err != nil {
			t.Fatal(err)
		}

		v, err = zetasql.
			NewValidation("SELECT @scalar_repeated AS scalar", v1beta1.NewParams(param)).
			ValidateQuery()
		if err != nil {
			t.Fatal(fmt.Errorf("scalar_repeated %s => %s error: %w", fType, zetasql.ZetaTypes[fType].Kind(), err))
		}

		for _, f := range v.Fields() {
			if f.Type.NativeType != string(fType) {
				t.Fatalf("expected %s, got %s", fType, f.Type)
			} else if !f.Repeated {
				t.Fatalf("expected repeated, got %v", f.Repeated)
			}
		}
	}
}

func TestValidation_ComplexFieldTypes(t *testing.T) {
	var complexTypes = map[bigquery.FieldType]any{
		bigquery.JSONFieldType:      map[string]any{"foo": "bar"},
		bigquery.RecordFieldType:    struct{ Foo string }{"bar"},
		bigquery.GeographyFieldType: `POINT(1 2)`,
		bigquery.IntervalFieldType:  `INTERVAL 1 DAY`,
	}

	for fType, value := range complexTypes {
		param, err := v1beta1.NewParamWithValue(
			v1beta1.NewField(
				"complex",
				v1beta1.NewDataType(fType),
			),
			value,
		)
		if err != nil {
			t.Fatal(err)
		}

		v, err := zetasql.
			NewValidation("SELECT @complex AS complex", v1beta1.NewParams(param)).
			ValidateQuery()
		if err != nil {
			t.Fatal(fmt.Errorf("complex %s => %s error: %w", fType, zetasql.ZetaTypes[fType].Kind(), err))
		}

		for _, f := range v.Fields() {
			if f.Type.NativeType != string(fType) {
				t.Fatalf("expected %s, got %s", fType, f.Type)
			} else if f.Repeated == true {
				t.Fatalf("expected not repeated, got %v", f.Repeated)
			}
		}
	}

	for fType, value := range complexTypes {
		param, err := v1beta1.NewParamWithValue(
			v1beta1.NewField(
				"complex",
				v1beta1.NewDataType(fType),
				v1beta1.WithFieldRepeated(true),
			),
			[]any{value},
		)
		if err != nil {
			t.Fatal(err)
		}

		v, err := zetasql.
			NewValidation("SELECT @complex AS complex", v1beta1.NewParams(param)).
			ValidateQuery()
		if err != nil {
			t.Fatal(fmt.Errorf("complex %s => %s error: %w", fType, zetasql.ZetaTypes[fType].Kind(), err))
		}

		for _, f := range v.Fields() {
			if f.Type.NativeType != string(fType) {
				t.Fatalf("expected %s, got %s", fType, f.Type)
			} else if f.Repeated != true {
				t.Fatalf("expected repeated, got %v", f.Repeated)
			}
		}
	}
}
