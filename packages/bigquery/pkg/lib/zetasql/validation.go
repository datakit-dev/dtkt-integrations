package zetasql

import (
	"fmt"
	"slices"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	"github.com/goccy/go-zetasql"
	"github.com/goccy/go-zetasql/ast"
	"github.com/goccy/go-zetasql/resolved_ast"
	zetatypes "github.com/goccy/go-zetasql/types"
)

var SkipErrors = []string{
	"Function not found: JSON_SET",
}

type Validation struct {
	valid    bool
	err      error
	sql      string
	params   v1beta1.Params
	fields   v1beta1.Fields
	accessed []*catalogv1beta1.CatalogPermission
}

func NewValidation(sql string, params v1beta1.Params) *Validation {
	return &Validation{
		sql:    sql,
		params: params,
	}
}

func IgnoreError(err error) bool {
	return slices.ContainsFunc(SkipErrors, func(skip string) bool {
		return strings.Contains(err.Error(), skip)
	})
}

func (v *Validation) Valid() bool {
	return v.valid
}

func (v *Validation) Error() string {
	if v.err != nil {
		return v.err.Error()
	}
	return ""
}

func (v *Validation) Accessed() []*catalogv1beta1.CatalogPermission {
	return v.accessed
}

func (v *Validation) Query() string {
	return v.sql
}

func (v *Validation) Dialect() string {
	return "GoogleSQL"
}

func (v *Validation) Fields() v1beta1.Fields {
	return v.fields
}

func (v *Validation) Params() v1beta1.Params {
	return v.params
}

func (v *Validation) ValidateQuery(perms ...*catalogv1beta1.CatalogPermission) (*Validation, error) {
	opt, err := newAnalyzerOptions()
	if err != nil {
		return v, err
	}

	for _, param := range v.params {
		if ztype, ok := ZetaTypes[bigquery.FieldType(param.Field.Type.NativeType)]; ok {
			if param.Field.Repeated {
				ztype, err = zetatypes.NewArrayType(ztype)
				if err != nil {
					return v, err
				}

				err = opt.AddQueryParameter(param.Field.Name, ztype)
				if err != nil {
					return v, err
				}
			} else {
				err = opt.AddQueryParameter(param.Field.Name, ztype)
				if err != nil {
					return v, err
				}
			}
		} else {
			return v, fmt.Errorf("unsupported parameter type: %s", param.Field.Type.NativeType)
		}
	}

	stmt, err := zetasql.ParseStatement(v.sql, opt.ParserOptions())
	if err != nil {
		return v, err
	}

	root, err := NewRootCatalog(perms)
	if err != nil {
		return v, err
	}

	if err := v.resolveAccessed(stmt, root); err != nil {
		return v, err
	}

	v.sql = zetasql.Unparse(stmt)
	out, err := zetasql.AnalyzeStatementFromParserAST(v.sql, stmt, root, opt)
	if err != nil {
		return v, err
	}

	err = v.resolveFields(out.Statement(), stmt.Kind() == ast.QueryStatement, root)
	if err != nil {
		return v, err
	}

	v.valid = true
	return v, nil
}

func (v *Validation) resolveAccessed(stmt ast.StatementNode, root *Catalog) error {
	return ast.Walk(stmt, func(n ast.Node) error {
		switch n := n.(type) {
		case *ast.PathExpressionNode:
			children := n.Names()

			var names []string
			switch len(children) {
			case 1:
				names = strings.Split(children[0].Name(), ".")
			case 2:
				schName1 := children[0].Name()
				tblName1 := children[1].Name()
				names = []string{schName1, tblName1}

				for catName, schMap := range root.tblLookup {
					for schName2, tblMap := range schMap {
						for tblName2 := range tblMap {
							if strings.EqualFold(schName1, schName2) && strings.EqualFold(tblName1, tblName2) {
								names = []string{catName, schName1, tblName1}
								break
							}
						}
					}
				}
			case 3:
				names = []string{children[0].Name(), children[1].Name(), children[2].Name()}
			}

			if len(names) == 3 {
				var (
					catName string
					schName string
					tblName string
					catPerm *catalogv1beta1.CatalogPermission
					schPerm *catalogv1beta1.SchemaPermission
					tblPerm *catalogv1beta1.TablePermission
				)

				for _, perm := range root.catLookup {
					if strings.EqualFold(perm.Name, names[0]) || strings.EqualFold(perm.Alias, names[0]) {
						catName = perm.Name
						catPerm = v1beta1.NewCatalogPermission(perm.Name, perm.Alias)
						break
					}
				}

				if catPerm == nil {
					return fmt.Errorf("catalog not found: %s", names[0])
				}

				for _, perm := range root.schLookup[catName] {
					if strings.EqualFold(perm.Name, names[1]) || strings.EqualFold(perm.Alias, names[1]) {
						schName = perm.Name
						schPerm = v1beta1.NewSchemaPermission(perm.Name, perm.Alias)
						catPerm.Schemas = append(catPerm.Schemas, schPerm)
						break
					}
				}

				if schPerm == nil {
					return fmt.Errorf("schema not found: %s.%s", catName, names[1])
				}

				for _, perm := range root.tblLookup[catName][schName] {
					if strings.EqualFold(perm.Name, names[2]) || strings.EqualFold(perm.Alias, names[2]) {
						tblName = perm.Name
						tblPerm = v1beta1.NewTablePermission(perm.Name, perm.Alias)
						schPerm.Tables = append(schPerm.Tables, tblPerm)
						break
					}
				}

				if tblPerm == nil {
					return fmt.Errorf("table not found: %s.%s.%s", catName, schName, names[2])
				}

				v.accessed = append(v.accessed, catPerm)

				switch len(children) {
				case 1:
					children[0].SetName(fmt.Sprintf("%s.%s.%s", catName, schName, tblName))
				case 2:
					children[0].SetName(schName)
					children[1].SetName(tblName)
				case 3:
					if slices.Index(names, children[0].Name()) == 0 {
						children[0].SetName(catName)
					}
					if slices.Index(names, children[1].Name()) == 1 {
						children[1].SetName(schName)
					}
					if slices.Index(names, children[2].Name()) == 2 {
						children[2].SetName(tblName)
					}
				}
			}
		}

		return nil
	})
}

func (v *Validation) resolveFields(stmt resolved_ast.StatementNode, isQuery bool, root *Catalog) error {
	return resolved_ast.Walk(stmt, func(n resolved_ast.Node) error {
		switch n := n.(type) {
		case *resolved_ast.OutputColumnNode:
			if isQuery {
				if field, err := NewField(n.Name(), n.Column().Type()); err != nil {
					return err
				} else {
					v.fields = append(v.fields, field)
				}
			}
		case *resolved_ast.TableScanNode:
			for _, zcol := range n.ColumnList() {
				var (
					tblParts []string
					colPerm  *catalogv1beta1.ColumnPermission
				)

			outer:
				for _, catPerm := range v.accessed {
					for _, schPerm := range catPerm.Schemas {
						for _, tblPerm := range schPerm.Tables {
							if strings.EqualFold(tblPerm.Name, zcol.TableName()) || strings.EqualFold(tblPerm.Alias, zcol.TableName()) {
								tblParts = []string{catPerm.Name, schPerm.Name, tblPerm.Name}
								if l1, ok := root.colLookup[catPerm.Name]; ok {
									if l2, ok := l1[schPerm.Name]; ok {
										if l3, ok := l2[tblPerm.Name]; ok {
											for _, perm := range l3 {
												if strings.EqualFold(perm.Name, zcol.Name()) || strings.EqualFold(perm.Alias, zcol.Name()) {
													colPerm = v1beta1.NewColumnPermission(perm.Name, perm.Alias, perm.Type)
													tblPerm.Columns = append(tblPerm.Columns, colPerm)
													break outer
												}
											}
										}
									}
								}
							}
						}
					}
				}

				if len(tblParts) == 0 {
					tblParts = strings.Split(zcol.TableName(), ".")
				}

				if colPerm == nil {
					return fmt.Errorf("column %s not found in: %s", zcol.Name(), strings.Join(tblParts, "."))
				}
			}
		}
		return nil
	})
}
