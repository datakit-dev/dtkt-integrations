package v1beta1_test

// import (
// 	"testing"

// 	"github.com/stretchr/testify/suite"
// )

// func TestQueryPage(t *testing.T) {
// 	suite.Run(t, new(QueryPageTestSuite))
// }

// type QueryPageTestSuite struct {
// 	suite.Suite
// }

// func (s *QueryPageTestSuite) TestQueryPage() {
// 	jobID := "job_id"
// 	pageToken := "page token"
// 	prev := newQueryPage(jobID, pageToken, nil)
// 	p := newQueryPage(jobID, pageToken, prev)

// 	s.Equal(jobID, p.JobID)
// 	s.Equal(pageToken, p.PageToken)
// 	s.Equal(prev, p.PrevPage)
// }

// func (s *QueryPageTestSuite) TestEncodeQueryPage() {
// 	jobID := "job_id"
// 	pageToken := "page token"

// 	p := newQueryPage(jobID, pageToken, nil)
// 	encoded := p.encode()

// 	s.Equal("eyJqb2JfaWQiOiJqb2JfaWQiLCJwYWdlX3Rva2VuIjoicGFnZSB0b2tlbiJ9", encoded)
// }

// func (s *QueryPageTestSuite) TestDecodeQueryPage() {
// 	encoded := "eyJqb2JfaWQiOiJqb2JfaWQiLCJwYWdlX3Rva2VuIjoicGFnZSB0b2tlbiJ9"

// 	p, err := decodeResultPage[queryPage](encoded)
// 	s.NoError(err)
// 	s.Equal("job_id", p.JobID)
// 	s.Equal("page token", p.PageToken)
// 	s.Nil(p.PrevPage)
// }

// func (s *QueryPageTestSuite) TestEncodeQueryPageWithPrev() {
// 	jobID := "job_id"
// 	pageToken := "page token"

// 	prev := newQueryPage(jobID, pageToken, nil)

// 	p := newQueryPage(jobID, pageToken, prev)
// 	encoded := p.encode()

// 	s.Equal("eyJqb2JfaWQiOiJqb2JfaWQiLCJwYWdlX3Rva2VuIjoicGFnZSB0b2tlbiIsInByZXZfcGFnZSI6eyJqb2JfaWQiOiJqb2JfaWQiLCJwYWdlX3Rva2VuIjoicGFnZSB0b2tlbiJ9fQ==", encoded)

// 	p, err := decodeResultPage[queryPage](encoded)
// 	s.NoError(err)
// 	s.Equal("job_id", p.JobID)
// 	s.Equal("page token", p.PageToken)
// 	s.NotNil(p.PrevPage)
// 	s.Equal("job_id", p.PrevPage.JobID)
// 	s.Equal("page token", p.PrevPage.PageToken)
// 	s.Nil(p.PrevPage.PrevPage)

// }
