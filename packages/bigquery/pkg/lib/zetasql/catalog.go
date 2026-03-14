package zetasql

import (
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	zetatypes "github.com/goccy/go-zetasql/types"
)

type Catalog struct {
	*zetatypes.SimpleCatalog
	catLookup map[string]*catalogv1beta1.CatalogPermission
	schLookup map[string]map[string]*catalogv1beta1.SchemaPermission
	tblLookup map[string]map[string]map[string]*catalogv1beta1.TablePermission
	colLookup map[string]map[string]map[string]map[string]*catalogv1beta1.ColumnPermission
}

func NewRootCatalog(accessible []*catalogv1beta1.CatalogPermission) (*Catalog, error) {
	var (
		zcatMap   = map[string]*zetatypes.SimpleCatalog{}
		catLookup = map[string]*catalogv1beta1.CatalogPermission{}
		schLookup = map[string]map[string]*catalogv1beta1.SchemaPermission{}
		tblLookup = map[string]map[string]map[string]*catalogv1beta1.TablePermission{}
		colLookup = map[string]map[string]map[string]map[string]*catalogv1beta1.ColumnPermission{}
	)

	root := zetatypes.NewSimpleCatalog("root")
	root.AddZetaSQLBuiltinFunctions(nil)

	for _, cat := range accessible {
		catLookup[cat.Name] = cat
		schLookup[cat.Name] = map[string]*catalogv1beta1.SchemaPermission{}
		tblLookup[cat.Name] = map[string]map[string]*catalogv1beta1.TablePermission{}
		colLookup[cat.Name] = map[string]map[string]map[string]*catalogv1beta1.ColumnPermission{}

		var (
			zcat *zetatypes.SimpleCatalog
			ok   bool
		)
		if zcat, ok = zcatMap[cat.Name]; !ok {
			zcat = zetatypes.NewSimpleCatalog(cat.Name)
			root.AddCatalog(zcat)
			zcatMap[cat.Name] = zcat
		}

		for _, sch := range cat.Schemas {
			schLookup[cat.Name][sch.Name] = sch
			tblLookup[cat.Name][sch.Name] = map[string]*catalogv1beta1.TablePermission{}
			colLookup[cat.Name][sch.Name] = map[string]map[string]*catalogv1beta1.ColumnPermission{}
			zsch := zetatypes.NewSimpleCatalog(sch.Name)

			for _, tbl := range sch.Tables {
				tblLookup[cat.Name][sch.Name][tbl.Name] = tbl
				colLookup[cat.Name][sch.Name][tbl.Name] = map[string]*catalogv1beta1.ColumnPermission{}

				zcols := make([]zetatypes.Column, 0, len(tbl.Columns))
				for _, col := range tbl.Columns {
					colLookup[cat.Name][sch.Name][tbl.Name][col.Name] = col

					var (
						ctype zetatypes.Type
						ok    bool
					)

					if ctype, ok = ZetaTypes[bigquery.FieldType(col.Type)]; !ok {
						return nil, fmt.Errorf("unsupported column type: %s.%s.%s %s(%s)",
							cat.Name, sch.Name, tbl.Name, col.Name, col.Type,
						)
					}

					zcol := zetatypes.NewSimpleColumn(tbl.Name, col.Name, ctype)
					zcols = append(zcols, zcol)
				}
				zsch.AddTable(zetatypes.NewSimpleTable(tbl.Name, zcols))
			}
			zcat.AddCatalogWithName(sch.Name, zsch)
		}
	}

	return &Catalog{
		root,
		catLookup,
		schLookup,
		tblLookup,
		colLookup,
	}, nil
}

func (c *Catalog) FindTable(path []string) (zetatypes.Table, error) {
	tbl, err := c.SimpleCatalog.FindTable(path)
	if err != nil {
		switch len(path) {
		case 1:
			path = strings.Split(path[0], ".")
		case 2:
			for catName := range c.catLookup {
				for schName := range c.schLookup[catName] {
					for tblName := range c.tblLookup[catName][path[0]] {
						if strings.EqualFold(path[0], schName) && strings.EqualFold(path[1], tblName) {
							path = []string{catName, schName, tblName}
							break
						}
					}
				}
			}
		}

		if len(path) == 3 {
			cats, err := c.Catalogs()
			if err != nil {
				return nil, err
			}

			for _, cat := range cats {
				if strings.EqualFold(cat.FullName(), path[0]) {
					return cat.FindTable(path[1:])
				}
			}
		}
	}

	return tbl, err
}
