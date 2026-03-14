package v1beta1

import (
	"testing"

	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib/zetasql"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestQuery(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

type QueryTestSuite struct {
	suite.Suite
}

func (s *QueryTestSuite) TestQueryValidate() {
	v, err := zetasql.
		NewValidation("SELECT 1", nil).
		ValidateQuery()

	require.NoError(s.T(), err)
	require.True(s.T(), v.Valid())
}

func (s *QueryTestSuite) TestQueryFields() {
	v, err := zetasql.
		NewValidation(`SELECT 1 AS x, "foo" AS y`, nil).
		ValidateQuery()

	require.NoError(s.T(), err)
	require.True(s.T(), v.Valid())

	fields := v.Fields()
	require.Len(s.T(), fields, 2)

	x := fields[0]
	require.Equal(s.T(), "x", x.Name)
	require.Equal(s.T(), "INTEGER", x.Type.NativeType)
	require.False(s.T(), x.Nullable)
	require.False(s.T(), x.Repeated)

	y := fields[1]
	require.Equal(s.T(), "y", y.Name)
	require.Equal(s.T(), "STRING", y.Type.NativeType)
	require.False(s.T(), y.Nullable)
	require.False(s.T(), y.Repeated)
}

func (s *QueryTestSuite) TestQueryFieldsFromParams() {
	var params v1beta1.Params

	p, err := v1beta1.NewParamWithValue(
		v1beta1.NewField("x",
			v1beta1.NewDataType("INTEGER"),
			v1beta1.WithFieldNullable(true),
		),
		1,
	)
	require.NoError(s.T(), err)
	params = append(params, p)

	p, err = v1beta1.NewParamWithValue(
		v1beta1.NewField("y",
			v1beta1.NewDataType("STRING"),
			v1beta1.WithFieldNullable(true),
		),
		"foo",
	)
	require.NoError(s.T(), err)
	params = append(params, p)

	v, err := zetasql.
		NewValidation(`SELECT @x AS x, @y AS y`, params).
		ValidateQuery()
	require.NoError(s.T(), err)
	require.True(s.T(), v.Valid())

	fields := v.Fields()
	require.Len(s.T(), fields, 2)

	x := fields[0]
	require.Equal(s.T(), "x", x.Name)
	require.Equal(s.T(), "INTEGER", x.Type.NativeType)
	// require.True(s.T(), x.Nullable)
	require.False(s.T(), x.Repeated)

	y := fields[1]
	require.Equal(s.T(), "y", y.Name)
	require.Equal(s.T(), "STRING", y.Type.NativeType)
	// require.True(s.T(), y.Nullable)
	require.False(s.T(), y.Repeated)
}

// func (s *QueryTestSuite) TestQueryParamsNullValue() {
// 	for _, t := range s.client.DataTypes() {
// 		var param property.Param
// 		if t == string(bigquery.RecordFieldType) {
// 			param = property.Param{
// 				Name: "x",
// 				Type: t,
// 				Fields: []property.Param{
// 					{Name: "y", Type: string(bigquery.IntegerFieldType)},
// 				},
// 				Value: nil,
// 			}
// 			q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 			require.Error(s.T(), err)
// 			require.False(s.T(), q.Valid())
// 			require.Error(s.T(), q.Err())
// 		} else {
// 			param = property.Param{Name: "x", Type: t, Value: nil}
// 			q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 			require.NoError(s.T(), err)
// 			require.True(s.T(), q.Valid())
// 			require.NoError(s.T(), q.Err())
// 		}
// 	}
// }

// func (s *QueryTestSuite) TestQueryParamsDefaultValue() {
// 	for _, t := range s.client.DataTypes() {
// 		var param property.Param
// 		if td, ok := typeDefault[t]; ok {
// 			if t == string(bigquery.RecordFieldType) {
// 				param = property.Param{
// 					Name: "x",
// 					Type: t,
// 					Fields: []property.Param{
// 						{Name: "y", Type: string(bigquery.IntegerFieldType)},
// 					},
// 					Value: td,
// 				}
// 				q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 				require.Error(s.T(), err)
// 				require.False(s.T(), q.Valid())
// 				require.Error(s.T(), q.Err())
// 			} else {
// 				param = property.Param{Name: "x", Type: t, Value: td}
// 				q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 				require.NoError(s.T(), err)
// 				require.True(s.T(), q.Valid())
// 				require.NoError(s.T(), q.Err())
// 			}
// 		}
// 	}
// }

// func (s *QueryTestSuite) TestQueryParamsRepeatedEmptyValues() {
// 	for _, t := range s.client.DataTypes() {
// 		var param property.Param
// 		if t == string(bigquery.RecordFieldType) {
// 			param = property.Param{
// 				Name:     "x",
// 				Type:     t,
// 				Repeated: true,
// 				Fields: []property.Param{
// 					{Name: "y", Type: string(bigquery.IntegerFieldType)},
// 				},
// 				Values: []any{},
// 			}
// 			q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 			require.Error(s.T(), err)
// 			require.False(s.T(), q.Valid())
// 			require.Error(s.T(), q.Err())
// 		} else {
// 			param = property.Param{Name: "x", Type: t, Repeated: true, Values: []any{}}
// 			q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 			require.NoError(s.T(), err)
// 			require.True(s.T(), q.Valid())
// 			require.NoError(s.T(), q.Err())
// 		}
// 	}
// }

// func (s *QueryTestSuite) TestQueryParamsRepeatedDefaultValues() {
// 	for _, t := range s.client.DataTypes() {
// 		var param property.Param
// 		if td, ok := typeDefault[t]; ok {
// 			if t == string(bigquery.RecordFieldType) {
// 				param = property.Param{
// 					Name:     "x",
// 					Type:     t,
// 					Repeated: true,
// 					Fields: []property.Param{
// 						{Name: "y", Type: string(bigquery.IntegerFieldType)},
// 					},
// 					Values: []any{
// 						td,
// 					},
// 				}
// 				q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 				require.Error(s.T(), err)
// 				require.False(s.T(), q.Valid())
// 				require.Error(s.T(), q.Err())
// 			} else {
// 				param = property.Param{Name: "x", Type: t, Repeated: true, Values: []any{td}}
// 				q, err := s.client.ValidateQuery(context.Background(), "SELECT @x AS x", []property.Param{param})
// 				require.NoError(s.T(), err)
// 				require.True(s.T(), q.Valid())
// 				require.NoError(s.T(), q.Err())
// 			}
// 		}
// 	}
// }

// func (s *QueryTestSuite) TestQueryResults() {
// 	q, err := s.client.GetQueryResults(context.Background(), "SELECT 1 AS x", nil, 0, nil)
// 	require.NoError(s.T(), err)
// 	require.True(s.T(), q.Valid())
// 	require.NoError(s.T(), q.Err())

// 	rows := q.Rows()
// 	require.Len(s.T(), rows, 1)

// 	row := rows[0]
// 	require.Len(s.T(), row, 1)

// 	value, ok := row["x"]
// 	require.True(s.T(), ok)
// 	require.Equal(s.T(), int64(1), value)
// }
